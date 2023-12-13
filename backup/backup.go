package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pkg/errors"
	"github.com/quic-go/quic-go/http3"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	dirDateTimeFormat = "20060102"
	maxSingleDirSize  = 18 << 30
	ErrorEventID      = 99
	BackupOutPath     = "/carfile/titan"
	StorageAPI        = "https://api-storage.container1.titannet.io"
)

var log = logging.Logger("backup")

var (
	waitInterval   = time.Minute * 5
	backupInterval = time.Hour * 1
)

type Downloader struct {
	lk         sync.Mutex
	schedulers []*statistics.Scheduler

	JobQueue []*model.Asset
	dirSize  map[string]int64
	token    string
}

func newDownloader(token string, scheduler []*statistics.Scheduler) *Downloader {
	return &Downloader{
		JobQueue:   make([]*model.Asset, 0),
		dirSize:    make(map[string]int64),
		schedulers: scheduler,
		token:      token,
	}
}

func (d *Downloader) GetJobs() []*model.Asset {
	d.lk.Lock()
	defer d.lk.Unlock()

	var out []*model.Asset
	copy(out, d.JobQueue)
	d.JobQueue = nil

	return out
}

func (d *Downloader) Push(jobs []*model.Asset) {
	d.lk.Lock()
	defer d.lk.Unlock()

	d.JobQueue = append(d.JobQueue, jobs...)
}

func (d *Downloader) create(ctx context.Context, job *model.Asset) (*model.Asset, error) {
	dir := job.EndTime.Format(dirDateTimeFormat)

	outPath, err := d.getOutPath(dir)
	if err != nil {
		return nil, err
	}

	err = d.download(ctx, outPath, job.Cid, job.TotalSize)
	if err != nil {
		log.Errorf("download CARFile %s: %v", job.Cid, err)
		job.Event = ErrorEventID
		return job, err
	}

	job.Path = outPath
	return job, nil
}

func (d *Downloader) download(ctx context.Context, outPath, cid string, size int64) error {
	var outErr error

	for _, scheduler := range d.schedulers {
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

			d.lk.Lock()
			d.dirSize[outPath] += size
			d.lk.Unlock()

			outErr = nil

			log.Infof("Successfully download CARFile %s.\n", outPath)
			return nil
		}
	}

	return outErr
}

func (d *Downloader) async() {
	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			assets, err := getJobs()
			if err != nil {
				log.Errorf("get jobs: %v", err)
				continue
			}
			d.Push(assets)
			ticker.Reset(backupInterval)
		}
	}

}

func (d *Downloader) run() {
	for {
		if len(d.JobQueue) == 0 {
			time.Sleep(waitInterval)
			continue
		}

		var todo []*model.Asset

		jobs := d.GetJobs()
		for _, job := range jobs {
			j, err := d.create(context.Background(), job)
			if err != nil {
				log.Errorf("download: %v", err)
			}
			todo = append(todo, j)
		}

		err := pushResult(d.token, todo)
		if err != nil {
			log.Errorf("push result: %v", err)
		}
	}
}

func (d *Downloader) createOrGetSize(dir string) (int64, error) {
	if !fileutil.Exist(dir) {
		return 0, os.Mkdir(dir, 0775)
	}

	d.lk.Lock()
	defer d.lk.Unlock()

	if size, ok := d.dirSize[dir]; ok {
		return size, nil
	}

	size, err := getDirSize(dir)
	if err != nil {
		return 0, err
	}
	d.dirSize[dir] = size

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

func (d *Downloader) getOutPath(dir string) (string, error) {
	var outPath string

	for c := 'a'; c < 'z'; c++ {
		outPath = filepath.Join(BackupOutPath, fmt.Sprintf("%s%c", dir, c))
		size, err := d.createOrGetSize(outPath)
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

func request(url, cid string, token *types.Token) (io.ReadCloser, error) {
	var scheme string
	if !strings.HasPrefix(url, "http") {
		scheme = "https://"
	}

	endpoint := fmt.Sprintf("%s%s/ipfs/%s?format=car", scheme, url, cid)

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

func pushResult(token string, jobs []*model.Asset) error {
	data, err := json.Marshal(jobs)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/backup_result", StorageAPI)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d %v", resp.StatusCode, resp.Status)
	}

	log.Infof("success")
	return nil
}

func getJobs() ([]*model.Asset, error) {
	url := fmt.Sprintf("%s/backup_assets", StorageAPI)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d %v", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out []*model.Asset
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
