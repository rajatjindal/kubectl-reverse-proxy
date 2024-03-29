package dashboard

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type server struct {
	port            string
	stopCh          chan struct{}
	client          *http.Client
	rvaddr          string
	instanceMetrics []*InstantMetricData

	sync.Mutex
}

func New(dashport string, rvaddr string) *server {
	return &server{
		rvaddr: rvaddr,
		port:   dashport,
		stopCh: make(chan struct{}),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("access-control-allow-methods", "OPTIONS, GET, POST, PATCH, PUT, HEAD, DELETE")
		w.Header().Set("access-control-allow-headers", "authorization, content-type")

		return
	}

	if r.URL.Path == "/metrics" {
		s.ServeMetrics(w, r)
		return
	}

	s.ServeIndex(w, r)
}

func (s *server) Start() {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", s.port),
		Handler:        s,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		server.ListenAndServe()
	}()

	go s.startCollector()
}
