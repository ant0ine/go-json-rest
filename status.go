package rest

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Status contains stats and status information. It is returned by GetStatus.
// These information can be made available as an API endpoint, see the "status"
// example to install the following status route.
// GET /.status returns something like:
//
//     {
//       "Pid": 21732,
//       "UpTime": "1m15.926272s",
//       "UpTimeSec": 75.926272,
//       "Time": "2013-03-04 08:00:27.152986 +0000 UTC",
//       "TimeUnix": 1362384027,
//       "StatusCodeCount": {
//         "200": 53,
//         "404": 11
//       },
//       "TotalCount": 64,
//       "TotalResponseTime": "16.777ms",
//       "TotalResponseTimeSec": 0.016777,
//       "AverageResponseTime": "262.14us",
//       "AverageResponseTimeSec": 0.00026214
//     }
type Status struct {
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

func (s *statusService) update(statusCode int, responseTime *time.Duration) {
	s.lock.Lock()
	s.responseCounts[fmt.Sprintf("%d", statusCode)]++
	s.totalResponseTime = s.totalResponseTime.Add(*responseTime)
	s.lock.Unlock()
}

func (s *statusService) getStatus() *Status {
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

	return &Status{
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
}

// GetStatus returns a Status object. EnableStatusService must be true.
func (rh *ResourceHandler) GetStatus() *Status {
	return rh.statusService.getStatus()
}

// The middleware function.
func (rh *ResourceHandler) statusWrapper(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		if rh.statusService != nil {
			rh.statusService.update(
				r.Env["statusCode"].(int),
				r.Env["elapsedTime"].(*time.Duration),
			)
		}
	}
}
