package rest

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// statusMiddleware keeps track of various stats about the processed requests.
// It depends on request.Env["STATUS_CODE"] and request.Env["ELAPSED_TIME"]
type statusMiddleware struct {
	lock              sync.RWMutex
	start             time.Time
	pid               int
	responseCounts    map[string]int
	totalResponseTime time.Time
}

func newStatusMiddleware() *statusMiddleware {
	return &statusMiddleware{
		start:             time.Now(),
		pid:               os.Getpid(),
		responseCounts:    map[string]int{},
		totalResponseTime: time.Time{},
	}
}

func (mw *statusMiddleware) update(statusCode int, responseTime *time.Duration) {
	mw.lock.Lock()
	mw.responseCounts[fmt.Sprintf("%d", statusCode)]++
	mw.totalResponseTime = mw.totalResponseTime.Add(*responseTime)
	mw.lock.Unlock()
}

func (mw *statusMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		mw.update(
			r.Env["STATUS_CODE"].(int),
			r.Env["ELAPSED_TIME"].(*time.Duration),
		)
	}
}

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

func (mw *statusMiddleware) getStatus() *Status {

	mw.lock.RLock()

	now := time.Now()

	uptime := now.Sub(mw.start)

	totalCount := 0
	for _, count := range mw.responseCounts {
		totalCount += count
	}

	totalResponseTime := mw.totalResponseTime.Sub(time.Time{})

	averageResponseTime := time.Duration(0)
	if totalCount > 0 {
		avgNs := int64(totalResponseTime) / int64(totalCount)
		averageResponseTime = time.Duration(avgNs)
	}

	status := &Status{
		Pid:                    mw.pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        mw.responseCounts,
		TotalCount:             totalCount,
		TotalResponseTime:      totalResponseTime.String(),
		TotalResponseTimeSec:   totalResponseTime.Seconds(),
		AverageResponseTime:    averageResponseTime.String(),
		AverageResponseTimeSec: averageResponseTime.Seconds(),
	}

	mw.lock.RUnlock()

	return status
}
