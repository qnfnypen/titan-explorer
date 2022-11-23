package api

import "fmt"

func (s *Server) AddStatisticsTask() {
	s.cron.AddFunc("1 * * * * *", DefaultStatistics)
}

func DefaultStatistics() {
	fmt.Println("implement me")
}
