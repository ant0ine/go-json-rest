package rest

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type statusService struct {
	lock              sync.Mutex
	start             time.Time
	pid               int
	responseCounts    map[string]int
	totalResponseTime time.Time
}

func newStatusService() *statusService {
	return &statusService{
		start:             time.Now(),
		pid:               os.Getpid(),
		responseCounts:    map[string]int{},
		totalResponseTime: time.Time{},
	}
}

func (self *statusService) update(statusCode int, responseTime *time.Duration) {
	self.lock.Lock()
	self.responseCounts[fmt.Sprintf("%d", statusCode)]++
	self.totalResponseTime = self.totalResponseTime.Add(*responseTime)
	self.lock.Unlock()
}

type status struct {
	Pid                    int
	UpTime                 string
	UpTimeSec              float64
	Time                   string
	TimeUnix               int64
	StatusCodeCount        map[string]int
	TotalCount             int
	TotalResponseTime      string
	TotalResponseTimeSec   float64
	AverageResponseTime    string
	AverageResponseTimeSec float64
}

func (self *statusService) getStatus(w *ResponseWriter, r *Request) {

	now := time.Now()

	uptime := now.Sub(self.start)

	totalCount := 0
	for _, count := range self.responseCounts {
		totalCount += count
	}

	totalResponseTime := self.totalResponseTime.Sub(time.Time{})

	averageResponseTime := time.Duration(0)
	if totalCount > 0 {
		avgNs := int64(totalResponseTime) / int64(totalCount)
		averageResponseTime = time.Duration(avgNs)
	}

	st := &status{
		Pid:                    self.pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        self.responseCounts,
		TotalCount:             totalCount,
		TotalResponseTime:      totalResponseTime.String(),
		TotalResponseTimeSec:   totalResponseTime.Seconds(),
		AverageResponseTime:    averageResponseTime.String(),
		AverageResponseTimeSec: averageResponseTime.Seconds(),
	}

	err := w.WriteJson(st)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
