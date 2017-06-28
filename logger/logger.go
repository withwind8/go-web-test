package logger

import (
	"log"
	"net/http"
	"time"

	"github.com/withwind8/middleware"
)

type logger struct {
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next func()) {
	start := time.Now()
	next()
	ww := w.(*middleware.ResponseWriter)
	log.Printf("%v %v %v %v %v %v", r.Method, r.URL.Path, ww.Status(), ww.Size(), time.Since(start), w.Header().Get("Content-Type"))
}

func New() *logger {
	return &logger{}
}
