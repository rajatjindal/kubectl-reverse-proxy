package dashboard

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

//go:embed index.html
var index []byte

func (s *server) ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html")
	w.Write(index)
}

func (s *server) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	s.Lock()
	ims := s.instanceMetrics
	s.Unlock()

	raw, err := json.Marshal(toDashboard(ims))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(raw)
}
