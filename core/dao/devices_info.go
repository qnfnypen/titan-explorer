package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"strings"
	"time"
)

var (
	tableNameDeviceInfo = "device_info"
)

type ActiveInfoOut struct {
	DeviceId     string `db:"device_id" json:"device_id"`
	ActiveStatus string `db:"active_status" json:"active_status"`
	Secret       string `db:"secret" json:"secret"`
}

func GetDeviceInfoList(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*model.DeviceInfo, int64, error) {
	var args []interface{}
	where := `WHERE device_id <> ''`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.BindStatus != "" {
		where += ` AND bind_status = ?`
		args = append(args, cond.BindStatus)
	}
	if cond.ActiveStatus < 10 {
		where += ` AND active_status = ?`
		args = append(args, cond.ActiveStatus)
	}
	if cond.UserID != "" {
		where += ` AND user_id = ?`
		args = append(args, cond.UserID)
	}
	if cond.DeviceStatus != "" && cond.DeviceStatus != "allDevices" {
		where += ` AND device_status = ?`
		args = append(args, cond.DeviceStatus)
	}
	if cond.IpLocation != "" && cond.IpLocation != "0" {
		where += ` AND ip_location = ?`
		args = append(args, cond.IpLocation)
	}
	if cond.NodeType > 0 {
		where += ` AND node_type = ?`
		args = append(args, cond.NodeType)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 500
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.DeviceInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s ORDER BY device_rank LIMIT %d OFFSET %d`, tableNameDeviceInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return HandleIpInfo(out), total, err
}

func GetDeviceActiveInfoList(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*ActiveInfoOut, int64, error) {
	var args []interface{}
	where := `WHERE a.device_id <> ''`
	if cond.DeviceID != "" {
		where += ` AND a.device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.BindStatus != "" {
		where += ` AND b.bind_status = ?`
		args = append(args, cond.BindStatus)
	}
	if cond.ActiveStatus < 10 {
		where += ` AND b.active_status = ?`
		args = append(args, cond.ActiveStatus)
	}
	if cond.UserID != "" {
		where += ` AND a.user_id = ?`
		args = append(args, cond.UserID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 500
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*)  FROM %s a LEFT JOIN %s b on a.device_id = b.device_id %s`, tableNameApplicationResult, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}
	var out []*ActiveInfoOut
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT a.device_id,IFNULL(0,b.active_status) active_status,a.secret FROM %s a LEFT JOIN %s b on a.device_id = b.device_id %s ORDER BY device_rank LIMIT %d OFFSET %d`, tableNameApplicationResult, tableNameDeviceInfo, where, limit, offset,
	), args...)

	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func HandleIpInfo(in []*model.DeviceInfo) []*model.DeviceInfo {
	for _, deviceInfo := range in {
		eIp := strings.Split(deviceInfo.ExternalIp, ".")
		if len(eIp) > 3 {
			deviceInfo.ExternalIp = eIp[0] + "." + "xxx" + "." + "xxx" + "." + eIp[3]
		}
		iIp := strings.Split(deviceInfo.InternalIp, ".")
		if len(iIp) > 3 {
			deviceInfo.InternalIp = iIp[0] + "." + "xxx" + "." + "xxx" + "." + iIp[3]
		}
	}
	return in
}

func GetDeviceInfoListByKey(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*model.DeviceInfo, int64, error) {
	var args []interface{}
	where := `WHERE device_id <> '' AND active_status = 1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.UserID != "" {
		where += ` AND user_id = ?`
		args = append(args, cond.UserID)
	} else {
		where += ` AND user_id <> ''`
	}
	if cond.DeviceStatus != "" && cond.DeviceStatus != "allDevices" {
		where += ` AND device_status = ?`
		args = append(args, cond.DeviceStatus)
	}
	if cond.IpLocation != "" && cond.IpLocation != "0" {
		where += ` AND ip_location = ?`
		args = append(args, cond.IpLocation)
	}
	if cond.NodeType > 0 {
		where += ` AND node_type = ?`
		args = append(args, cond.NodeType)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.DeviceInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s ORDER BY device_rank LIMIT %d OFFSET %d`, tableNameDeviceInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return HandleIpInfo(out), total, err
}

func HandleMapList(ctx *gin.Context, deviceIfo *model.DeviceInfo) *model.DeviceInfo {
	if ctx.GetHeader("Lang") == "en" {
		continents := strings.Split(deviceIfo.IpLocation, "-")
		continent := "Asia"
		if len(continents) > 1 {
			switch continents[0] {
			case "亚洲":
				continent = "Asia"
				deviceIfo.IpProvince = GetProvinceName(ctx.Request.Context(), deviceIfo.IpProvince)
				deviceIfo.IpCity = GetCityName(ctx.Request.Context(), deviceIfo.IpCity)
			case "欧洲":
				continent = "Europe"
			case "非洲":
				continent = "Africa"
			case "大洋洲":
				continent = "Oceania"
			case "南极洲":
				continent = "Antarctica"
			case "北美洲":
				continent = "North America"
			case "南美洲":
				continent = "South America"
			default:
				continent = "Asia"

			}
			deviceIfo.IpLocation = continent
			deviceIfo.IpLocation += "-" + GetCountryName(ctx.Request.Context(), deviceIfo.IpCountry)
			deviceIfo.IpLocation += "-" + deviceIfo.IpProvince
			deviceIfo.IpLocation += "-" + deviceIfo.IpCity
		}

	}
	return deviceIfo
}

func HandleMapInfo(ctx *gin.Context, in []*model.DeviceInfo) []map[string]interface{} {
	type MapObject map[string]interface{}
	var mapInfoOut []map[string]interface{}
	mapLocationExit := make(map[float64]float64)
	for _, m := range in {
		if m.DeviceStatus == "offline" {
			continue
		}
		Latitude, ok := mapLocationExit[m.Longitude]
		mapLocationExit[m.Longitude] = m.Latitude
		if ok && Latitude == m.Latitude {
			m.Latitude += utils.RandFloat64() / 10000
			m.Longitude += utils.RandFloat64() / 10000
		}
		HandleMapList(ctx, m)
		mapInfoOut = append(mapInfoOut, MapObject{
			"name":     m.IpCity,
			"nodeType": m.NodeType,
			"ip":       m.ExternalIp,
			"value":    []float64{m.Latitude, m.Longitude},
		})

	}
	return mapInfoOut

}

func GetDeviceInfoByID(ctx context.Context, deviceID string) (*model.DeviceInfo, error) {
	var out model.DeviceInfo
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id = ?`, tableNameDeviceInfo), deviceID,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func UpdateUserDeviceInfo(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET user_id = :user_id, updated_at = now(),bind_status = :bind_status WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func UpdateDownloadCount(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET download_count = :download_count, total_upload = :total_upload,updated_at = now() WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func UpdateValidateCount(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET total_upload = :total_upload,updated_at = now() WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func UpdateTotalDownload(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET total_download = :total_download, updated_at = now() WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func UpdateDeviceName(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = now(),device_name = :device_name WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func UpdateDeviceStatus(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = now(),device_status = :device_status, device_status_code = :device_status_code WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func BulkUpsertDeviceInfo(ctx context.Context, deviceInfos []*model.DeviceInfo) error {
	statement := upsertDeviceInfoStatement()
	_, err := DB.NamedExecContext(ctx, statement, deviceInfos)
	return err
}

func BulkUpdateDeviceInfo(ctx context.Context, deviceInfos []*model.DeviceInfo) error {
	for _, device := range deviceInfos {
		_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
			`UPDATE %s SET today_online_time = :today_online_time, today_profit = :today_profit,
				yesterday_profit = :yesterday_profit, seven_days_profit = :seven_days_profit, month_profit = :month_profit, 
				updated_at = now() WHERE device_id = :device_id`, tableNameDeviceInfo),
			device)
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertDeviceInfoStatement() string {
	insertStatement := fmt.Sprintf(
		`INSERT INTO %s (
                	device_id, node_type, device_name, user_id, system_version,  active_status,network_info, external_ip, internal_ip, ip_location,
                	ip_country, ip_province, ip_city, latitude, longitude, mac_location, cpu_usage, memory_usage, cpu_cores, memory, disk_usage, disk_space,
                	device_status, device_status_code, io_system, online_time, today_online_time, today_profit, yesterday_profit, seven_days_profit, month_profit,
                	cumulative_profit, bandwidth_up, bandwidth_down,download_traffic,upload_traffic, created_at, updated_at, bound_at,cache_count,retrieval_count
                	)
				VALUES (
					:device_id, :node_type, :device_name, :user_id,  :system_version, :active_status,:network_info, :external_ip, :internal_ip, :ip_location,
					:ip_country, :ip_province, :ip_city, :latitude, :longitude, :mac_location,:cpu_usage, :memory_usage, :cpu_cores, :memory, :disk_usage, :disk_space,
					:device_status, :device_status_code, :io_system, :online_time, :today_online_time, :today_profit,:yesterday_profit, :seven_days_profit, :month_profit,
					:cumulative_profit, :bandwidth_up, :bandwidth_down,:download_traffic,:upload_traffic, now(), now(),:bound_at,:cache_count,:retrieval_count
				)`, tableNameDeviceInfo,
	)
	updateStatement := ` ON DUPLICATE KEY UPDATE node_type = VALUES(node_type),  device_name = VALUES(device_name),active_status = VALUES(active_status),
				system_version = VALUES(system_version), network_info = VALUES(network_info), cumulative_profit = VALUES(cumulative_profit),
				external_ip = VALUES(external_ip), internal_ip = VALUES(internal_ip), ip_location = VALUES(ip_location), ip_country = VALUES(ip_country), 
				ip_province = VALUES(ip_province), ip_city = VALUES(ip_city),latitude = VALUES(latitude), longitude = VALUES(longitude), mac_location = VALUES(mac_location),
				online_time = VALUES(online_time),cpu_usage = VALUES(cpu_usage), cpu_cores = VALUES(cpu_cores),  memory_usage = VALUES(memory_usage), memory = VALUES(memory),
				disk_usage = VALUES(disk_usage), disk_space = VALUES(disk_space), device_status = VALUES(device_status), device_status_code = VALUES(device_status_code) ,io_system = VALUES(io_system), bandwidth_up = VALUES(bandwidth_up),
				bandwidth_down = VALUES(bandwidth_down),download_traffic = VALUES(download_traffic),upload_traffic = VALUES(upload_traffic), updated_at = now(),bound_at = VALUES(bound_at),cache_count = VALUES(cache_count),retrieval_count = VALUES(retrieval_count)`
	return insertStatement + updateStatement
}

func GetAllAreaFromDeviceInfo(ctx context.Context) ([]string, error) {
	queryStatement := fmt.Sprintf(`SELECT ip_location FROM %s GROUP BY ip_location;`, tableNameDeviceInfo)
	var out []string
	err := DB.SelectContext(ctx, &out, queryStatement)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func SumFullNodeInfoFromDeviceInfo(ctx context.Context) (*model.FullNodeInfo, error) {
	queryStatement := fmt.Sprintf(`SELECT count( device_id ) AS total_node_count ,  SUM(IF(node_type = 1, 1, 0)) AS edge_count, 
       SUM(IF(node_type = 2, 1, 0)) AS candidate_count,sum(cache_count) as t_upstream_file_count,
       count(device_status = 'online' or null) as online_node_count,
       SUM(IF(node_type = 3, 1, 0)) AS validator_count, ROUND(count(device_status = 'online' or null)/count( device_id )*100,2) AS t_node_online_ratio,
       ROUND(SUM( disk_space),4) AS total_storage, ROUND(SUM( disk_usage*disk_space/100),4) AS storage_used, 
       ROUND(SUM(bandwidth_up),2) AS total_upstream_bandwidth, ROUND(SUM(bandwidth_down),2) AS total_downstream_bandwidth FROM %s where active_status = 1;`, tableNameDeviceInfo)

	var out model.FullNodeInfo
	if err := DB.QueryRowxContext(ctx, queryStatement).StructScan(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UserDeviceProfile struct {
	CumulativeProfit float64 `json:"cumulative_profit" db:"cumulative_profit"`
	YesterdayProfit  float64 `json:"yesterday_profit" db:"yesterday_profit"`
	TodayProfit      float64 `json:"today_profit" db:"today_profit"`
	SevenDaysProfit  float64 `json:"seven_days_profit" db:"seven_days_profit"`
	MonthProfit      float64 `json:"month_profit" db:"month_profit"`
	NodeType         *int    `json:"node_type" db:"node_type"`
	TotalNum         int64   `json:"total_num" db:"total_num"`
	OnlineNum        int64   `json:"online_num" db:"online_num"`
	OfflineNum       int64   `json:"offline_num" db:"offline_num"`
	AbnormalNum      int64   `json:"abnormal_num" db:"abnormal_num"`
	TotalBandwidth   float64 `json:"total_bandwidth" db:"total_bandwidth"`
}

func CountUserDeviceInfo(ctx context.Context, userID string) (*UserDeviceProfile, error) {
	queryStatement := fmt.Sprintf(`SELECT COALESCE(sum(cumulative_profit),0) as cumulative_profit, COALESCE(sum(yesterday_profit),0) as yesterday_profit, 
COALESCE(sum(today_profit),0) as today_profit,node_type, COALESCE(sum(seven_days_profit),0) as seven_days_profit, COALESCE(sum(month_profit),0) as month_profit, count(*) as total_num, 
count(IF(device_status = 'online', 1, NULL)) as online_num ,count(IF(device_status = 'offline', 1, NULL)) as offline_num, 
count(IF(device_status = 'abnormal', 1, NULL)) as abnormal_num, COALESCE(sum(bandwidth_up),0) as total_bandwidth from %s where user_id = ? and active_status = 1;`, tableNameDeviceInfo)

	var out UserDeviceProfile
	if err := DB.QueryRowxContext(ctx, queryStatement, userID).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func RankDeviceInfo(ctx context.Context) error {
	tx := DB.MustBegin()
	defer tx.Rollback()
	tx.MustExec("SET @r=0;")
	queryStatement := fmt.Sprintf(`UPDATE %s SET device_rank= @r:= (@r+1) ORDER BY device_status DESC, node_type DESC;`, tableNameDeviceInfo)
	_, err := tx.ExecContext(ctx, queryStatement)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GenerateInactiveNodeRecords(ctx context.Context, t time.Time) error {
	var inactiveNodeIds []model.DeviceInfo
	query := fmt.Sprintf("SELECT * FROM %s where updated_at < ?", tableNameDeviceInfo)
	err := DB.SelectContext(ctx, &inactiveNodeIds, query, t)
	if err != nil {
		return err
	}

	var inactiveNodes []*model.DeviceInfoHour
	insertRecordStatement := fmt.Sprintf("SELECT * FROM %s WHERE  device_id = ? ORDER BY time DESC limit 1", tableNameDeviceInfoHour)
	for _, deviceInfo := range inactiveNodeIds {
		newDIH := model.DeviceInfoHour{}
		err = DB.Get(&newDIH, insertRecordStatement, deviceInfo.DeviceID)
		if err != nil {
			log.Errorf("get inactive node last record,%s: %v", deviceInfo.DeviceID, err)
			continue
		}
		newDIH.CreatedAt = time.Now()
		newDIH.UpdatedAt = time.Now()
		newDIH.Time = t
		newDIH.OnlineTime = deviceInfo.OnlineTime
		newDIH.DiskUsage = deviceInfo.DiskUsage
		newDIH.DiskSpace = deviceInfo.DiskSpace
		newDIH.BlockCount = deviceInfo.CacheCount
		newDIH.RetrievalCount = deviceInfo.RetrievalCount
		newDIH.UserID = deviceInfo.UserID
		newDIH.UpstreamTraffic = deviceInfo.UploadTraffic
		newDIH.DownstreamTraffic = deviceInfo.DownloadTraffic
		inactiveNodes = append(inactiveNodes, &newDIH)
	}

	return BulkUpsertDeviceInfoHours(ctx, inactiveNodes)
}

func GetDeviceInfo(ctx context.Context, deviceId string) model.DeviceInfo {
	var deviceInfo model.DeviceInfo
	query := fmt.Sprintf("SELECT user_id FROM %s where device_id = '%s'", tableNameDeviceInfo, deviceId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&deviceInfo)
	if err != nil {
		log.Errorf("getDeviceInfo %v", err)
	}
	return deviceInfo
}

func GetDeviceInfoById(ctx context.Context, deviceId string) model.DeviceInfo {
	var deviceInfo model.DeviceInfo
	query := fmt.Sprintf("SELECT * FROM %s where device_id = '%s'", tableNameDeviceInfo, deviceId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&deviceInfo)
	if err != nil {
		log.Errorf("getDeviceInfo %v", err)
	}
	eIp := strings.Split(deviceInfo.ExternalIp, ".")
	if len(eIp) > 3 {
		deviceInfo.ExternalIp = eIp[0] + "." + "xxx" + "." + "xxx" + "." + eIp[3]
	}
	iIp := strings.Split(deviceInfo.InternalIp, ".")
	if len(iIp) > 3 {
		deviceInfo.InternalIp = iIp[0] + "." + "xxx" + "." + "xxx" + "." + iIp[3]
	}
	return deviceInfo
}

func GetNodesInfo(ctx context.Context, option QueryOption) (int64, []model.NodesInfo, error) {
	where := `WHERE device_id <> '' AND active_status = 1`
	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}
	var total int64
	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(distinct(user_id)) FROM %s %s`, tableNameDeviceInfo, where,
	))
	if err != nil {
		return 0, nil, err
	}
	var nodeInfo []model.NodesInfo
	query := fmt.Sprintf("SELECT node_type,user_id,COUNT(device_id) AS node_count,ROUND(sum(disk_space) ,2) as disk_space,ROUND(SUM(bandwidth_up),2) as bandwidth_up FROM %s %s GROUP BY user_id ORDER BY node_count DESC LIMIT %d OFFSET %d",
		tableNameDeviceInfo, where, limit, offset)
	err = DB.SelectContext(ctx, &nodeInfo, query)
	if err != nil {
		log.Errorf("GetNodesInfo %v", err)
		return 0, nil, err
	}
	return total, nodeInfo, nil
}

func GetIdIfExit(ctx context.Context, nodeId string) bool {
	var out model.DeviceInfo
	query := fmt.Sprintf(`SELECT * FROM %s where device_id = '%s';`, tableNameDeviceInfo, nodeId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Errorf("GetIdIfExit err:%v", err)
		return false
	}
	return true
}

func DeleteDeviceInfoHourHistory(ctx context.Context, before time.Time) error {
	statement := fmt.Sprintf(`DELETE FROM %s where created_at < ?`, tableNameDeviceInfoHour)
	_, err := DB.ExecContext(ctx, statement, before)
	return err
}
