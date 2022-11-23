package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func (t *DeviceTask) DeviceInfoGetFromRpc(url string, DeviceID string) (DeviceInfo model.DeviceInfo, err error) {
	var data RpcDevice
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.GetDevicesInfo"
	song["id"] = 3
	song["params"] = []string{DeviceID}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debug(string(respBytes))
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if GUpdate {
		var hourDaily model.HourDaily
		hourDaily.Time = GTime
		hourDaily.DiskUsage = data.Result.DiskUsage
		hourDaily.DeviceID = data.Result.DeviceID
		hourDaily.PkgLossRatio = data.Result.PkgLossRatio
		hourDaily.HourIncome = data.Result.TodayProfit
		//data.Result.TodayOnlineTime = data.Result.TodayOnlineTime
		hourDaily.OnlineTime = data.Result.OnlineTime
		hourDaily.Latency = data.Result.Latency
		hourDaily.DiskUsage = data.Result.DiskUsage
		_, ok := t.DeviceIDAndUserId[hourDaily.DeviceID]
		if ok {
			hourDaily.UserID = t.DeviceIDAndUserId[hourDaily.DeviceID]
		}
		err = TransferData(hourDaily)
		if err != nil {
			log.Error(err.Error())
		}
	}
	return data.Result, nil
}

func CidInfoGetFromRpc(url string, DeviceID string) error {
	var data RpcTask
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.GetDownloadInfo"
	song["id"] = 3
	song["params"] = []string{DeviceID}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	var dataSave model.TaskInfo
	if len(data.Result) > 0 {
		for _, taskOne := range data.Result {
			dataSave.Cid = taskOne.Cid
			dataSave.DeviceID = taskOne.DeviceId
			dataSave.FileSize = taskOne.FileSize
			dataSave.Price = taskOne.Reward
			dataSave.Time = taskOne.TimeDone
			dataSave.BandwidthUp = fmt.Sprintf("%f", taskOne.BandwidthUp)
			dataSave.Status = "已完成"
			err = SaveTaskInfo(dataSave)
			if err != nil {
				log.Error(err.Error())
				continue
			}
		}
	}
	return nil
}

func AllMinerInfoGetFromRpc(url string) {
	var data AllMinerInfo
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.StateNetwork"
	song["id"] = 3
	song["params"] = []string{}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	//defer client.CloseIdleConnections()
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return
	}
	AllM = data
	return
}

func (t *DeviceTask) SaveDeviceInfo(url string, Df string) error {
	data, err := t.DeviceInfoGetFromRpc(url, Df)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if data.DeviceID == "" {
		log.Info("empty response from rpc", Df)
		return nil
	}

	var DeviceInfoOld model.DeviceInfo
	data.SevenDaysProfit = 0
	data.MonthProfit = 0
	data.YesterdayIncome = 0

	ctx := context.Background()
	old, err := dao.GetDeviceInfoByID(ctx, data.DeviceID)
	if err != nil {
		log.Errorf("get device info by id: %v", err)
		return err
	}

	if old == nil {
		err = dao.CreateDeviceInfo(ctx, &data)
		if err != nil {
			log.Errorf("create device info: %v", err)
			return err
		}
	}

	data.ID = DeviceInfoOld.ID
	data.UpdatedAt = time.Now()
	err = dao.UpdateDeviceInfo(ctx, &data)
	if err != nil {
		log.Errorf("update device info: %v", err)
		return err
	}
	return nil
}

func SaveTaskInfo(data model.TaskInfo) error {
	if data.DeviceID == "" {
		log.Info("empty response from rpc", data)
		return nil
	}

	ctx := context.Background()
	old, err := dao.GetTaskInfoByTime(ctx, data.DeviceID, data.Cid, data.Time)
	if err != nil {
		log.Errorf("get task info by time: %v", err)
		return err
	}

	if old == nil {
		data.CreatedAt = time.Now()
		return dao.CreateTaskInfo(ctx, &data)
	}

	data.ID = old.ID
	data.UpdatedAt = time.Now()
	return dao.UpdateTaskInfo(ctx, &data)
}

func TransferData(hourDaily model.HourDaily) error {
	if hourDaily.DeviceID == "" {
		return nil
	}

	hourDaily.UpdatedAt = time.Now()
	ctx := context.Background()
	old, err := dao.GetHourDailyByTime(ctx, hourDaily.DeviceID, hourDaily.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		hourDaily.CreatedAt = time.Now()
		return dao.CreateHourDaily(ctx, &hourDaily)
	}

	hourDaily.ID = old.ID
	return dao.UpdateHourDaily(ctx, &hourDaily)
}

func (t *DeviceTask) SavaIncomeDailyInfo(daily model.IncomeDaily) error {
	if daily.DeviceID == "" {
		return nil
	}

	daily.UpdatedAt = time.Now()
	_, ok := t.DeviceIDAndUserId[daily.DeviceID]
	if ok {
		daily.UserID = t.DeviceIDAndUserId[daily.DeviceID]
	}

	ctx := context.Background()
	old, err := dao.GetIncomeDailyByTime(ctx, daily.DeviceID, daily.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		daily.CreatedAt = time.Now()
		return dao.CreateIncomeDaily(ctx, &daily)
	}

	daily.ID = old.ID
	return dao.UpdateIncomeDaily(ctx, &daily)
}

func (t *DeviceTask) FormatIncomeDailyList(DeviceID string) {
	timeNow := time.Now().Format("2006-01-02")
	DateFrom := timeNow + " 00:00:00"
	DateTo := timeNow + " 23:59:59"
	sqlClause := fmt.Sprintf("select user_id,date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", DeviceID, DateFrom, DateTo)
	datas, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, data := range datas {
		var InPage model.IncomeDaily
		InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		InPage.DiskUsage = Str2Float64(data["disk_usage"])
		InPage.NatRatio = Str2Float64(data["nat_ratio"])
		InPage.Income = Str2Float64(data["hour_income"])
		InPage.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		InPage.Latency = Str2Float64(data["latency"])
		InPage.DeviceID = DeviceID
		InPage.UserID = data["user_id"]
		t.SavaIncomeDailyInfo(InPage)
	}
	return
}

func (t *DeviceTask) CountDataByUser(userId string) {
	dd, _ := time.ParseDuration("-24h")
	timeBase := time.Now().Add(dd * 1).Format("2006-01-02")
	DateFrom := timeBase + " 00:00:00"
	DateTo := timeBase + " 23:59:59"
	sqlClause := fmt.Sprintf("select user_id, sum(income) as income from income_daily "+
		"where  time>='%s' and time<='%s' and user_id='%s' group by user_id;", DateFrom, DateTo, userId)
	dataBase, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, data := range dataBase {
		var InPage model.IncomeDaily
		InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		InPage.DiskUsage = Str2Float64(data["disk_usage"])
		InPage.NatRatio = Str2Float64(data["nat_ratio"])
		InPage.Income = Str2Float64(data["hour_income"])
		InPage.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		InPage.Latency = Str2Float64(data["latency"])
		InPage.UserID = data["user_id"]
		t.SavaIncomeDailyInfo(InPage)
	}
	return
}

func (t *DeviceTask) UpdateYesTodayIncome(DeviceID string) error {
	dd, _ := time.ParseDuration("-24h")
	timeBase := time.Now().Add(dd * 1).Format("2006-01-02")
	timeNow := time.Now().Format("2006-01-02")
	DateFrom := timeBase + " 00:00:00"
	DateTo := timeBase + " 23:59:59"
	dataY := QueryDataByDate(DeviceID, DateFrom, DateTo)
	timeBase = time.Now().Add(dd * 6).Format("2006-01-02")
	DateFrom = timeBase + " 00:00:00"
	DateTo = timeNow + " 23:59:59"
	dataS := QueryDataByDate(DeviceID, DateFrom, DateTo)
	timeBase = time.Now().Add(dd * 29).Format("2006-01-02")
	DateFrom = timeBase + " 00:00:00"
	dataM := QueryDataByDate(DeviceID, DateFrom, DateTo)
	dataA := QueryDataByDate(DeviceID, "", "")
	DateFrom = timeNow + " 00:00:00"
	DateTo = timeNow + " 23:59:59"
	dataT := QueryDataByDate(DeviceID, DateFrom, DateTo)
	var dataUpdate model.DeviceInfo
	dataUpdate.YesterdayIncome = 0
	dataUpdate.SevenDaysProfit = 0
	dataUpdate.MonthProfit = 0
	dataUpdate.CumuProfit = 0
	dataUpdate.TodayOnlineTime = 0
	dataUpdate.TodayProfit = 0
	if len(dataY) > 0 {
		dataUpdate.YesterdayIncome = Str2Float64(dataY["income"])
	}
	if len(dataS) > 0 {
		dataUpdate.SevenDaysProfit = Str2Float64(dataS["income"])
	}
	if len(dataM) > 0 {
		dataUpdate.MonthProfit = Str2Float64(dataM["income"])
	}
	if len(dataA) > 0 {
		dataUpdate.CumuProfit = Str2Float64(dataA["income"])
	}
	if len(dataT) > 0 {
		dataUpdate.TodayProfit = Str2Float64(dataT["income"])
		dataUpdate.TodayOnlineTime = Str2Float64(dataT["online_time"])
	}
	dataUpdate.UpdatedAt = time.Now()
	_, ok := t.DeviceIDAndUserId[DeviceID]
	if ok {
		dataUpdate.UserID = t.DeviceIDAndUserId[DeviceID]
	}
	//err := dao.DB.Save(&data).Error

	ctx := context.Background()
	old, err := dao.GetDeviceInfoByID(ctx, DeviceID)
	if err != nil {
		log.Errorf("get device info by id: %v", err)
		return err
	}

	if old == nil {
		dataUpdate.CreatedAt = time.Now()
		return dao.CreateDeviceInfo(ctx, &dataUpdate)
	}
	old.YesterdayIncome = dataUpdate.YesterdayIncome
	old.SevenDaysProfit = dataUpdate.SevenDaysProfit
	old.MonthProfit = dataUpdate.MonthProfit
	old.CumuProfit = dataUpdate.CumuProfit
	old.UpdatedAt = dataUpdate.UpdatedAt
	old.TodayOnlineTime = dataUpdate.TodayOnlineTime
	old.TodayProfit = dataUpdate.TodayProfit
	if dataUpdate.UserID != "" {
		old.UserID = dataUpdate.UserID
	}
	err = dao.UpdateDeviceInfo(ctx, old)
	if err != nil {
		log.Errorf("update device info: %v", err)
	}
	return nil
}

func QueryDataByDate(DeviceID, DateFrom, DateTo string) map[string]string {

	sqlClause := fmt.Sprintf("select sum(income) as income,online_time from income_daily "+
		"where  time>='%s' and time<='%s' and device_id='%s' group by user_id;", DateFrom, DateTo, DeviceID)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select sum(income) as income,online_time from income_daily "+
			"where device_id='%s' group by user_id;", DeviceID)
	}
	data, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	if len(data) > 0 {
		return data[0]
	}
	return nil
}

var (
	GDevice       *DeviceTask
	GWg           *sync.WaitGroup
	GUpdateTagNew string
	GUpdate       bool
	GUpdateTask   bool
	GTime         time.Time
)

type DeviceTask struct {
	Done              chan struct{}
	RunInterval       int64
	DeviceIDs         []string
	DeviceIDAndUserId map[string]string
}

func (t *DeviceTask) Initial(interval int64) {
	t.Done = make(chan struct{}, 1)
	t.RunInterval = interval
	t.DeviceIDAndUserId = make(map[string]string)
	t.GetDeviceIDs()
	today := time.Now().Format(TimeFormatYMD)
	GUpdateTagNew = today
	GUpdate = false
	GUpdateTask = false
}

func (t *DeviceTask) GetDeviceIDs() {
	list, _, err := dao.GetDeviceInfoList(context.Background(), &model.DeviceInfo{}, dao.QueryOption{})
	if err != nil {
		log.Errorf("get device inf list: %v", err)
		return
	}
	for _, DeviceID := range list {
		t.DeviceIDs = append(t.DeviceIDs, DeviceID.DeviceID)
		if DeviceID.UserID != "" && DeviceID.DeviceID != "" {
			t.DeviceIDAndUserId[DeviceID.DeviceID] = DeviceID.UserID
		}
	}
	return
}

func (t *DeviceTask) itemRun(url string) {
	log.Infof("start item run: %s", url)
	defer GWg.Done()
	ticker := time.Tick(time.Duration(t.RunInterval) * time.Second)
	for {
		select {
		case <-t.Done:
			log.Infof("device Run once loop end")
			return
		default:
		}

		nowMin := time.Now().Minute()
		if nowMin%10 == 0 {
			GTime = time.Now()
			GUpdate = true
		}

		//today := time.Now().Format(TimeFormatYMD)
		//if GUpdateTagNew == "" || GUpdateTagNew != today {
		//	GUpdate = true
		//	GUpdateTagNew = today
		//}
		for _, DeviceID := range t.DeviceIDs {
			err := t.SaveDeviceInfo(url, DeviceID)
			if err != nil {
				log.Infof("wrong msg %v", err)
				<-ticker
				continue
			}
			if GUpdate {
				// 定时任务更新每日设备参数信息
				t.FormatIncomeDailyList(DeviceID)
				// 定时任务更新统计收入信息
				t.UpdateYesTodayIncome(DeviceID)
				// 定时更新全网数据
				AllMinerInfoGetFromRpc(url)
				// 更新设备完成任务
				err := CidInfoGetFromRpc(url, DeviceID)
				if err != nil {
					log.Infof("wrong msg %v", err)
					<-ticker
					continue
				}
			}

		}
		GUpdate = false
		<-ticker
	}
}

func RunTask() {
	GDevice = &DeviceTask{}
	GDevice.Initial(60)
	GWg = &sync.WaitGroup{}
	GTime = time.Now()

	schedulers, err := dao.GetSchedulers(context.Background())
	if err != nil {
		log.Fatalf("get scheduler: %v", err)
	}

	if len(schedulers) == 0 {
		log.Fatalf("scheulers not found")
	}

	log.Infof("total scheduler: %d", len(schedulers))

	GWg.Add(1)
	go GDevice.itemRun(schedulers[0].Address)

	GWg.Wait()
	log.Debug("run loop end")
}
