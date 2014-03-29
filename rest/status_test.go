package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestStatus(t *testing.T) {

	handler := ResourceHandler{
		EnableStatusService: true,
	}
	handler.SetRoutes(
		&Route{"GET", "/r",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
		&Route{"GET", "/.status",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(handler.GetStatus())
			},
		},
	)

	// one request to the API
	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/r", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// check the status
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/.status", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	payload := map[string]interface{}{}

	err := recorded.DecodeJsonPayload(&payload)
	if err != nil {
		t.Fatal(err)
	}

	if payload["Pid"] == nil {
		t.Error("Expected a non nil Pid")
	}

	if payload["TotalCount"].(float64) != 1 {
		t.Errorf("TotalCount 1 Expected, got: %f", payload["TotalCount"].(float64))
	}

	if payload["StatusCodeCount"].(map[string]interface{})["200"].(float64) != 1 {
		t.Errorf("StatusCodeCount 200 1 Expected, got: %f", payload["StatusCodeCount"].(map[string]interface{})["200"].(float64))
	}
}
