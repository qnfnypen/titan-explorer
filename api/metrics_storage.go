package api

import (
	"context"
	"time"

	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/prometheus/client_golang/prometheus"
)

var (

	// 累计文件下载数
	View_TotalDownloadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_download_count",
		Help: "Total number of file downloads",
	})

	// 累计文件上传数
	View_TotalUploadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_upload_count",
		Help: "Total number of file uploads",
	})

	// 累计文件下载大小 (以 GB 记录)
	View_TotalDownloadSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_download_size_gb",
		Help: "Total size of downloaded files in GB",
	})

	// 累计文件上传大小 (以 GB 记录)
	View_TotalUploadSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_upload_size_gb",
		Help: "Total size of uploaded files in GB",
	})

	// 累计文件下载成功数
	View_TotalDownloadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_download_success_count",
		Help: "Total number of successful file downloads",
	})

	// 累计文件上传成功数
	View_TotalUploadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_upload_success_count",
		Help: "Total number of successful file uploads",
	})

	// 累计文件下载失败数
	View_TotalDownloadFailure = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_download_failure_count",
		Help: "Total number of failed file downloads",
	})

	// 累计文件上传失败数
	View_TotalUploadFailure = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_upload_failure_count",
		Help: "Total number of failed file uploads",
	})

	// 文件下载平均速度
	View_TotalDownloadAvgSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "download_avg_speed_mb_s",
		Help: "Average download speed in MB/s",
	})

	// 文件上传平均速度
	View_TotalUploadAvgSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upload_avg_speed_mb_s",
		Help: "Average upload speed in MB/s",
	})

	// 今日文件下载数
	View_TodayDownloadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_download_count",
		Help: "Today number of file downloads",
	})

	// 今日文件上传数
	View_TodayUploadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_upload_count",
		Help: "Today number of file uploads",
	})

	// 今日文件下载大小 (以 GB 记录)
	View_TodayDownloadSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_download_size_gb",
		Help: "Today size of downloaded files in GB",
	})

	// 今日文件上传大小 (以 GB 记录)
	View_TodayUploadSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_upload_size_gb",
		Help: "Today size of uploaded files in GB",
	})

	// 今日文件下载成功数
	View_TodayDownloadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_download_success_count",
		Help: "Today number of successful file downloads",
	})

	// 今日文件上传成功数
	View_TodayUploadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_upload_success_count",
		Help: "Today number of successful file uploads",
	})

	// 今日文件下载失败数
	View_TodayDownloadFailure = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_download_failure_count",
		Help: "Today number of failed file downloads",
	})

	// 今日文件上传失败数
	View_TodayUploadFailure = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_upload_failure_count",
		Help: "Today number of failed file uploads",
	})

	// 今日文件下载平均速度
	View_TodayDownloadAvgSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_download_avg_speed_mb_s",
		Help: "Today Average download speed in MB/s",
	})

	// 今日文件上传平均速度
	View_TodayUploadAvgSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "today_upload_avg_speed_mb_s",
		Help: "Today Average upload speed in MB/s",
	})
)

func init() {
	prometheus.MustRegister(View_TotalDownloadCount)
	prometheus.MustRegister(View_TotalUploadCount)
	prometheus.MustRegister(View_TotalDownloadSize)
	prometheus.MustRegister(View_TotalUploadSize)
	prometheus.MustRegister(View_TotalDownloadSuccess)
	prometheus.MustRegister(View_TotalUploadSuccess)
	prometheus.MustRegister(View_TotalDownloadFailure)
	prometheus.MustRegister(View_TotalUploadFailure)
	prometheus.MustRegister(View_TotalDownloadAvgSpeed)
	prometheus.MustRegister(View_TotalUploadAvgSpeed)

	prometheus.MustRegister(View_TodayDownloadCount)
	prometheus.MustRegister(View_TodayUploadCount)
	prometheus.MustRegister(View_TodayDownloadSize)
	prometheus.MustRegister(View_TodayUploadSize)
	prometheus.MustRegister(View_TodayDownloadSuccess)
	prometheus.MustRegister(View_TodayUploadSuccess)
	prometheus.MustRegister(View_TodayDownloadFailure)
	prometheus.MustRegister(View_TodayUploadFailure)
	prometheus.MustRegister(View_TodayDownloadAvgSpeed)
	prometheus.MustRegister(View_TodayUploadAvgSpeed)
}

func setStorageGatherer(ctx context.Context) {
	totalStats, err := dao.GetComprehensiveStatsInPeriod(ctx, 0, 0)
	if err != nil {
		log.Errorf("[gatherer] get total stats error: %s", err.Error())
	}
	if totalStats != nil {
		View_TotalDownloadCount.Set(float64(totalStats.TotalDownloads))
		View_TotalUploadCount.Set(float64(totalStats.TotalUploads))
		View_TotalDownloadSuccess.Set(float64(totalStats.DownloadSuccess))
		View_TotalUploadSuccess.Set(float64(totalStats.UploadSuccess))
		View_TotalDownloadFailure.Set(float64(totalStats.DownloadFailure))
		View_TotalUploadFailure.Set(float64(totalStats.UploadFailure))
		View_TotalDownloadSize.Set(float64(totalStats.DownloadSize) / (1024 * 1024 * 1024))
		View_TotalUploadSize.Set(float64(totalStats.UploadSize) / (1024 * 1024 * 1024))
		View_TotalDownloadAvgSpeed.Set(float64(totalStats.DownloadAvgSpeed))
		View_TotalUploadAvgSpeed.Set(float64(totalStats.UploadAvgSpeed))
	}

	beginToday := time.Now().Truncate(24 * time.Hour).Unix()
	todayStats, err := dao.GetComprehensiveStatsInPeriod(ctx, beginToday, 0)
	if err != nil {
		log.Errorf("[gatherer] get today stats error: %s", err.Error())
	}
	if todayStats != nil {
		View_TodayDownloadCount.Set(float64(todayStats.TotalDownloads))
		View_TodayUploadCount.Set(float64(todayStats.TotalUploads))
		View_TodayDownloadSuccess.Set(float64(todayStats.DownloadSuccess))
		View_TodayUploadSuccess.Set(float64(todayStats.UploadSuccess))
		View_TodayDownloadFailure.Set(float64(todayStats.DownloadFailure))
		View_TodayUploadFailure.Set(float64(todayStats.UploadFailure))
		View_TodayDownloadSize.Set(float64(todayStats.DownloadSize) / (1024 * 1024 * 1024))
		View_TodayUploadSize.Set(float64(todayStats.UploadSize) / (1024 * 1024 * 1024))
		View_TodayDownloadAvgSpeed.Set(float64(todayStats.DownloadAvgSpeed))
		View_TodayUploadAvgSpeed.Set(float64(todayStats.UploadAvgSpeed))
	}
}
