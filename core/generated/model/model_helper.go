package model

import "time"

func (u User) IsCorpUser() bool {
	return u.TenantID != ""
}

type AssetTrasnferDetails []*AssetTrasnferDetail

func (atds AssetTrasnferDetails) GroupByNodeAndState() *AssetTrasnferDetail {
	if len(atds) == 0 {
		return nil
	}

	var totalSize, peek, totalElaspedTime int64
	var firstCreatedAt time.Time = time.Now()
	for _, v := range atds {
		tpeek := v.Size / v.ElaspedTime
		if tpeek > peek {
			peek = tpeek
		}
		totalSize += v.Size
		totalElaspedTime += v.ElaspedTime
		if v.CreateAt.Before(firstCreatedAt) {
			firstCreatedAt = v.CreateAt
		}
	}

	ret := &AssetTrasnferDetail{
		TraceId:      atds[0].TraceId,
		NodeID:       atds[0].NodeID,
		State:        atds[0].State,
		TransferType: atds[0].TransferType,
		Peek:         peek,
		Size:         totalSize,
		ElaspedTime:  totalElaspedTime,
		CreateAt:     firstCreatedAt,
	}

	return ret
}
