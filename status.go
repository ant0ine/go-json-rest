package rest

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type status_service struct {
	lock                sync.Mutex
	start               time.Time
	pid                 int
	response_counts     map[string]int
	total_response_time time.Time
}

func new_status_service() *status_service {
	return &status_service{
		start:               time.Now(),
		pid:                 os.Getpid(),
		response_counts:     map[string]int{},
		total_response_time: time.Time{},
	}
}

func (self *status_service) update(status_code int, response_time *time.Duration) {
	self.lock.Lock()
	self.response_counts[fmt.Sprintf("%d", status_code)]++
	self.total_response_time = self.total_response_time.Add(*response_time)
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

func (self *status_service) get_status(w *ResponseWriter, r *Request) {

	now := time.Now()

	uptime := now.Sub(self.start)

	total_count := 0
	for _, count := range self.response_counts {
		total_count += count
	}

	total_response_time := self.total_response_time.Sub(time.Time{})

	average_response_time := time.Duration(0)
	if total_count > 0 {
		avg_ns := int64(total_response_time) / int64(total_count)
		average_response_time = time.Duration(avg_ns)
	}

	st := &status{
		Pid:                    self.pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        self.response_counts,
		TotalCount:             total_count,
		TotalResponseTime:      total_response_time.String(),
		TotalResponseTimeSec:   total_response_time.Seconds(),
		AverageResponseTime:    average_response_time.String(),
		AverageResponseTimeSec: average_response_time.Seconds(),
	}

	err := w.WriteJson(st)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
