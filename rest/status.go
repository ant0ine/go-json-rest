package rest

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// StatusMiddleware keeps track of various stats about the processed requests.
// It depends on request.Env["STATUS_CODE"] and request.Env["ELAPSED_TIME"],
// recorderMiddleware and timerMiddleware must be in the wrapped middlewares.
type StatusMiddleware struct {
	lock              sync.RWMutex
	start             time.Time
	pid               int
	responseCounts    map[string]int
	totalResponseTime time.Time
}

// MiddlewareFunc makes StatusMiddleware implement the Middleware interface.
func (mw *StatusMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	mw.start = time.Now()
	mw.pid = os.Getpid()
	mw.responseCounts = map[string]int{}
	mw.totalResponseTime = time.Time{}

	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		if r.Env["STATUS_CODE"] == nil {
			log.Fatal("StatusMiddleware: Env[\"STATUS_CODE\"] is nil, " +
				"RecorderMiddleware may not be in the wrapped Middlewares.")
		}
		statusCode := r.Env["STATUS_CODE"].(int)

		if r.Env["ELAPSED_TIME"] == nil {
			log.Fatal("StatusMiddleware: Env[\"ELAPSED_TIME\"] is nil, " +
				"TimerMiddleware may not be in the wrapped Middlewares.")
		}
		responseTime := r.Env["ELAPSED_TIME"].(*time.Duration)

		mw.lock.Lock()
		mw.responseCounts[fmt.Sprintf("%d", statusCode)]++
		mw.totalResponseTime = mw.totalResponseTime.Add(*responseTime)
		mw.lock.Unlock()
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

// GetStatus computes and returns a Status object based on the request informations accumulated
// since the start of the process.
func (mw *StatusMiddleware) GetStatus() *Status {

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
