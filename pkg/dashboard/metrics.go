package dashboard

import (
	"math"
	"time"

	dto "github.com/prometheus/client_model/go"
)

type Dataset struct {
	Label           string  `json:"label"`
	BorderColor     string  `json:"borderColor,omitempty"`
	BackgroundColor string  `json:"backgroundColor,omitempty"`
	Data            []int   `json:"data"`
	Fill            bool    `json:"fill"`
	Tension         float64 `json:"tension"`
}

type LineChart struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

type InstantMetricData struct {
	Timestamp             time.Time
	HealthyServersCount   int
	UnHealthyServersCount int
}

func processMetrics(mfMap map[string]*dto.MetricFamily, processorList ...processor) *InstantMetricData {
	im := &InstantMetricData{
		Timestamp: time.Now(),
	}

	for _, pFunc := range processorList {
		pFunc(im, mfMap)
	}

	return im
}

type processor func(im *InstantMetricData, mfMap map[string]*dto.MetricFamily)

func processHealthyServersCount(im *InstantMetricData, mfMap map[string]*dto.MetricFamily) {
	mf, exists := mfMap["caddy_reverse_proxy_upstreams_healthy"]
	if !exists {
		return
	}

	for _, m := range mf.Metric {
		if mf.GetType() != dto.MetricType_GAUGE {
			continue
		}

		if equalWithTolerane(1, *m.Gauge.Value) {
			im.HealthyServersCount++
		}

		if equalWithTolerane(0, m.Gauge.GetValue()) {
			im.UnHealthyServersCount++
		}
	}
}

func equalWithTolerane(a, b float64) bool {
	tolerance := 0.000001
	diff := math.Abs(a - b)
	return diff <= tolerance
}
