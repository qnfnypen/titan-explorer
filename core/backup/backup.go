package backup

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pkg/errors"
	"github.com/quic-go/quic-go/http3"
	"github.com/robfig/cron/v3"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultRequestLimit = 20
	dirDateTimeFormat   = "20060102"
	maxSingleDirSize    = 18 << 30
	ErrorEventID        = 99
)

var log = logging.Logger("backup")

type StorageBackup struct {
	schedulers []*statistics.Scheduler
	cron       *cron.Cron
	cfg        config.StorageBackupConfig

	lk      sync.RWMutex
	dirSize map[string]int64

	assetChan chan *model.Asset
}

func NewStorageBackup(cfg config.StorageBackupConfig, schedulers []*statistics.Scheduler) *StorageBackup {
	return &StorageBackup{
		cfg: cfg,
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLocation(time.Local),
		),
		schedulers: schedulers,
		dirSize:    make(map[string]int64),
		assetChan:  make(chan *model.Asset, 1),
	}
}

func (s *StorageBackup) Run() {
	go s.run()
	s.cron.AddFunc(s.cfg.Crontab, s.cronJob)
	s.cron.Start()
}

func (s *StorageBackup) cronJob() {
	if s.cfg.Disable {
		return
	}

	ctx := context.Background()
	var offset int64

Loop:
	assets, total, err := dao.GetAssetsByEmptyPath(ctx, defaultRequestLimit, offset)
	if err != nil {
		log.Errorf("GetAssertsByEmptyPath: %v", err)
		return
	}

	if len(assets) == 0 {
		return
	}

	offset += int64(len(assets))
	log.Debugf("loading assets %d/%d", offset, total)

	for _, assert := range assets {
		s.assetChan <- assert
	}

	if total > offset {
		goto Loop
	}

}

func (s *StorageBackup) run() {
	for {
		select {
		case asset := <-s.assetChan:
			if err := s.createAsset(context.Background(), asset); err != nil {
				log.Errorf("create assert :%v", err)
			}
		}
	}
}

func (s *StorageBackup) createAsset(ctx context.Context, asset *model.Asset) error {
	dir := asset.EndTime.Format(dirDateTimeFormat)

	outPath, err := s.getOutPath(dir)
	if err != nil {
		return err
	}

	err = s.download(ctx, outPath, asset.Cid, asset.TotalSize)
	if err != nil {
		log.Errorf("download CARFile %s: %v", asset.Cid, err)

		errx := dao.UpdateAssetEvent(ctx, asset.Cid, ErrorEventID)
		if errx != nil {
			log.Errorf("UpdateAssetEvent: %v", errx)
		}

		return err
	}

	if err = dao.UpdateAssetPath(ctx, asset.Cid, outPath); err != nil {
		log.Errorf("UpdateAssertPath: %v", err)
		return err
	}

	return nil
}

func (s *StorageBackup) getOutPath(dir string) (string, error) {
	var outPath string

	for c := 'a'; c < 'z'; c++ {
		outPath = filepath.Join(s.cfg.BackupPath, fmt.Sprintf("%s%c", dir, c))
		size, err := s.createOrGetSize(outPath)
		if err != nil {
			log.Errorf("createOrGetSize %s: %v", dir, err)
			return "", err
		}

		if size < maxSingleDirSize {
			break
		}
	}

	return outPath, nil
}

func (s *StorageBackup) download(ctx context.Context, outPath, cid string, size int64) error {
	var outErr error

	for _, scheduler := range s.schedulers {
		downloadInfos, err := scheduler.Api.GetCandidateDownloadInfos(ctx, cid)
		if err != nil {
			log.Errorf("GetCandidateDownloadInfos: %v", err)
			outErr = err
			continue
		}

		if len(downloadInfos) == 0 {
			outErr = errors.New(fmt.Sprintf("CARFile %s not found", cid))
			continue
		}

		for _, downloadInfo := range downloadInfos {
			reader, err := request(downloadInfo.Address, cid, downloadInfo.Tk)
			if err != nil {
				log.Errorf("download requeset: %v", err)
				outErr = err
				continue
			}

			file, err := os.Create(filepath.Join(outPath, cid+".car"))
			if err != nil {
				outErr = err
				return err
			}

			_, err = io.Copy(file, reader)
			if err != nil {
				outErr = err
				return err
			}

			s.lk.Lock()
			s.dirSize[outPath] += size
			s.lk.Unlock()

			outErr = nil

			log.Infof("Successfully download CARFile %s.\n", outPath)
			return nil
		}
	}

	return outErr
}

func (s *StorageBackup) createOrGetSize(dir string) (int64, error) {
	if !fileutil.Exist(dir) {
		return 0, os.Mkdir(dir, 0775)
	}

	s.lk.Lock()
	defer s.lk.Unlock()

	if size, ok := s.dirSize[dir]; ok {
		return size, nil
	}

	size, err := getDirSize(dir)
	if err != nil {
		return 0, err
	}
	s.dirSize[dir] = size

	return size, nil
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func request(url, cid string, token *types.Token) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("https://%s/ipfs/%s?format=car", url, cid)

	log.Debugf("endpoint: %s", endpoint)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	//req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	client := http.Client{
		Transport: &http3.RoundTripper{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http request: %d %v", resp.StatusCode, resp.Status)
	}

	return resp.Body, err
}
