package statistics

import (
	"context"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"github.com/ipfs/go-cid"
	"github.com/linguohua/titan/api"
	mh "github.com/multiformats/go-multihash"
	"time"
)

func (s *Statistic) CountCacheFiles() error {
	log.Info("start count cache files")
	start := time.Now()
	defer func() {
		log.Infof("count cache files, cost: %v", time.Since(start))
	}()

	ctx := context.Background()

	var (
		startTime, endTime time.Time
		sum                int64
	)

	lastEvent, err := dao.GetLastCacheEvent(ctx)
	if err != nil {
		log.Errorf("get last cache event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime = carbon.Now().StartOfDay().StartOfMinute().Carbon2Time()
	} else {
		startTime = lastEvent.Time
	}

	endTime = carbon.Time2Carbon(start).SubMinutes(start.Minute() % 5).StartOfMinute().Carbon2Time()
	req := api.ListCacheBlocksReq{
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Cursor:    0,
		Count:     1000,
	}

loop:
	resp, err := s.api.GetCacheBlockInfos(ctx, req)
	if err != nil {
		log.Errorf("api get cache block infos: %v", err)
		return err
	}

	var blockInfos []*model.BlockInfo
	for _, blockInfo := range resp.Data {
		blockInfos = append(blockInfos, toBlockInfo(blockInfo))
	}

	if len(blockInfos) > 0 {
		err = dao.CreateBlockInfo(ctx, blockInfos)
		if err != nil {
			log.Errorf("create block info: %v", err)
		}
	}

	sum += int64(len(resp.Data))
	req.Cursor += len(resp.Data)
	if sum < resp.Total {
		<-time.After(100 * time.Millisecond)
		goto loop
	}

	err = dao.TxStatisticDeviceBlocks(ctx, startTime, endTime)
	if err != nil {
		log.Errorf("statistics device blocks: %v", err)
		return err
	}

	return nil
}

func toBlockInfo(v interface{}) *model.BlockInfo {
	var blockInfo model.BlockInfo
	data, err := json.Marshal(v)
	if err != nil {
		log.Errorf("marshal device info: %v", err)
		return nil
	}

	err = json.Unmarshal(data, &blockInfo)
	if err != nil {
		return nil
	}

	blockInfo.CarfileCid = hashToCID(blockInfo.CarfileHash)

	return &blockInfo
}

func hashToCID(hashString string) string {
	multiphase, err := mh.FromHexString(hashString)
	if err != nil {
		return ""
	}
	cid := cid.NewCidV1(cid.Raw, multiphase)
	return cid.String()
}

func (s *Statistic) CountRetrieve() {

}
