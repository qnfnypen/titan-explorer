package main

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/viper"
	"log"
	"time"
)

const DefaultCommissionPercent = 5

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("reading config file: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unmarshaling config file: %v\n", err)
	}

	if err := dao.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	start := time.Now()

	ctx := context.Background()

	userRewardSum, err := SumAllUsersTotalReward(ctx)
	if err != nil {
		log.Fatalf("SumAllUsersReward: %v", err)
		return
	}

	kolLevels, err := dao.GetAllKOLLevels(ctx)
	if err != nil {
		log.Fatalf("GetAllKOLLevels: %v", err)
		return
	}

	referrerUserIdInUser, err := dao.GetAllUserReferrerUserId(ctx)
	if err != nil {
		log.Fatalf("GetAllUserReferrerUserId: %v", err)
		return
	}

	beforeDate := carbon.Yesterday().EndOfDay().String()
	maxUserRewardBefore, err := dao.GetAllUsersRewardBefore(ctx, beforeDate)
	if err != nil {
		log.Fatalf("GetAllUsersRewardBefore: %v", err)
		return
	}

	referralReward := make(map[string]float64)

	toLevelUpKOLs := make(map[string]*kolLevelRef)

	var updateUserRewards []*model.UserRewardDaily
	for _, userReward := range userRewardSum {
		if _, ok := maxUserRewardBefore[userReward.UserId]; ok {
			continue
		}

		userReward.UpdatedAt = start
		userReward.CreatedAt = start
		userReward.Time = carbon.Now().StartOfYear().StdTime()

		if _, ok := kolLevels[userReward.UserId]; ok {
			userReward.IsKOL = 1
		}

		if v, ok := referrerUserIdInUser[userReward.UserId]; ok {
			userReward.ReferrerUserId = v
		}

		if kol, ok := kolLevels[userReward.ReferrerUserId]; ok {
			userReward.IsReferrerKOL = 1
			userReward.KOLBonus = userReward.Reward * float64(kol.ChildrenBonusPercent) / float64(100)
			userReward.CommissionPercent = int64(kol.ParentCommissionPercent)

			cliReward := userReward.CliReward * DefaultCommissionPercent / 100
			appReward := userReward.AppReward * float64(kol.ParentCommissionPercent) / float64(100)

			userReward.ReferrerReward = cliReward + appReward
			referralReward[userReward.ReferrerUserId] += cliReward + appReward

			if _, exist := toLevelUpKOLs[userReward.ReferrerUserId]; !exist {
				toLevelUpKOLs[userReward.ReferrerUserId] = &kolLevelRef{
					CurrentLevel:               kol.Level,
					LevelUpNodeCountsThreshold: kol.DeviceThreshold,
				}
			}

			toLevelUpKOLs[userReward.ReferrerUserId].ReferralNodeCount += userReward.DeviceOnlineCount
		} else {
			userReward.IsReferrerKOL = 0
			reward := userReward.Reward * DefaultCommissionPercent / 100
			userReward.ReferrerReward = reward
			referralReward[userReward.ReferrerUserId] += reward
		}

		if _, ok := maxUserRewardBefore[userReward.UserId]; ok {
			continue
		}

		updateUserRewards = append(updateUserRewards, userReward)
	}

	var todos []*model.UserRewardDaily

	for i, u := range updateUserRewards {
		if reward, ok := referralReward[u.UserId]; ok {
			u.ReferralReward = reward
		}

		todos = append(todos, u)

		// Perform bulk insert when todos reaches a multiple of batchSize or at the end of updateUserRewards
		if len(todos)%1000 == 0 || i == len(updateUserRewards)-1 {
			if err := dao.BulkAddUserRewardDaily(context.Background(), todos); err != nil {
				log.Fatalf("BulkAddUserRewardDaily: %v", err)
			}
			todos = todos[:0] // Reset todos slice without reallocating memory
		}
	}

	log.Println("Success")
}

type kolLevelRef struct {
	CurrentLevel               int
	ReferralNodeCount          int64
	LevelUpNodeCountsThreshold int64
}

func unSizeVal(val float64) float64 {
	if val < 0 {
		return 0
	}
	return val
}

func SumAllUsersTotalReward(ctx context.Context) ([]*model.UserRewardDaily, error) {
	query := `select user_id, 
       ifnull(sum(cumulative_profit - today_profit ) ,0) as cumulative_reward, 
       ifnull(sum(cumulative_profit - today_profit ) ,0) as reward,
       sum(IF(app_type > 0,cumulative_profit - today_profit ,0)) as app_reward, 
       sum(IF(app_type = 0,cumulative_profit - today_profit ,0)) as cli_reward, 
       count(if(online_time >= 500, true, null)) as device_online_count,
       count(device_id) as total_device_count
		from device_info  where user_id <> '' GROUP BY user_id`

	var out []*model.UserRewardDaily
	err := dao.DB.SelectContext(ctx, &out, query)
	if err != nil {
		return nil, err
	}

	return out, nil
}
