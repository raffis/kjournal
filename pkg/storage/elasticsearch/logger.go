package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"k8s.io/klog/v2"
)

var _ elastictransport.Logger = &logger{}

type logger struct {
}

func (l *logger) LogRoundTrip(w *http.Request, r *http.Response, esErr error, ts time.Time, duration time.Duration) error {
	var b []byte
	if w.Body != nil {
		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			return err
		}

		b = body
	}

	klog.InfoS("elasticsearch roundtrip", "body", b, "method", w.Method, "uri", w.URL.String(), "err", esErr, "duration", duration, "responseCode", r.StatusCode)
	return nil

}

func (l *logger) RequestBodyEnabled() bool {
	return true
}

func (l *logger) ResponseBodyEnabled() bool {
	return false
}
