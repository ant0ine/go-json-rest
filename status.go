package rest

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

func (rh *ResourceHandler) statusWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// call the handler
		h(w, r)

		if rh.statusService != nil {
			rh.statusService.update(
				rh.env.getVar(r, "statusCode").(int),
				rh.env.getVar(r, "elapsedTime").(*time.Duration),
			)
		}
	}
}

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

func (s *statusService) getRoute() Route {
	return Route{
		HttpMethod: "GET",
		PathExp:    "/.status",
		Func: func(writer *ResponseWriter, request *Request) {
			s.getStatus(writer, request)
		},
	}
}

func (s *statusService) update(statusCode int, responseTime *time.Duration) {
	s.lock.Lock()
	s.responseCounts[fmt.Sprintf("%d", statusCode)]++
	s.totalResponseTime = s.totalResponseTime.Add(*responseTime)
	s.lock.Unlock()
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

func (s *statusService) getStatus(w *ResponseWriter, r *Request) {

	now := time.Now()

	uptime := now.Sub(s.start)

	totalCount := 0
	for _, count := range s.responseCounts {
		totalCount += count
	}

	totalResponseTime := s.totalResponseTime.Sub(time.Time{})

	averageResponseTime := time.Duration(0)
	if totalCount > 0 {
		avgNs := int64(totalResponseTime) / int64(totalCount)
		averageResponseTime = time.Duration(avgNs)
	}

	st := &status{
		Pid:                    s.pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        s.responseCounts,
		TotalCount:             totalCount,
		TotalResponseTime:      totalResponseTime.String(),
		TotalResponseTimeSec:   totalResponseTime.Seconds(),
		AverageResponseTime:    averageResponseTime.String(),
		AverageResponseTimeSec: averageResponseTime.Seconds(),
	}

	err := w.WriteJson(st)
	if err != nil {
		Error(w, err.Error(), 500)
	}
}
