package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zenoss/glog"
)

type responseLogRecord struct {
	StatusCode   int
	ResponseTime *time.Duration
	HttpMethod   string
	RequestURI   string
}

func (self *ResourceHandler) logResponseRecord(record *responseLogRecord) {
	if self.EnableLogAsJson {
		b, err := json.Marshal(record)
		if err != nil {
			panic(err)
		}
		glog.V(1).Infof("%s", b)
	} else {
		glog.V(1).Infof("%d %v %s %s",
			record.StatusCode,
			record.ResponseTime,
			record.HttpMethod,
			record.RequestURI,
		)
	}
}

func (self *ResourceHandler) logWrapper(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// call the handler
		h(w, r)

		self.logResponseRecord(&responseLogRecord{
			self.env.getVar(r, "statusCode").(int),
			self.env.getVar(r, "elapsedTime").(*time.Duration),
			r.Method,
			r.URL.RequestURI(),
		})
	}
}
