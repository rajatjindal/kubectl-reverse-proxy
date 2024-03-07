package dashboard

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"
)

func (s *server) startCollector() error {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			err := s.scrapeOnce()
			if err != nil {
				fmt.Printf("error scraping data from reverse proxy endpoint%v\n", err)
			}
		case <-s.stopCh:
			return nil
		}
	}
}

func (s *server) scrapeOnce() error {
	resp, err := s.client.Get(s.rvaddr)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	imData, err := s.parseRawMetrics(raw)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()
	s.instanceMetrics = append(s.instanceMetrics, imData)

	return nil
}

func (s *server) parseRawMetrics(raw []byte) (*InstantMetricData, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	imData := processMetrics(mf,
		processHealthyServersCount,
	)

	return imData, nil
}

type Dashboard struct {
	HealthyServersPanel LineChart `json:"healthyServersPanel"`
}

func toDashboard(ims []*InstantMetricData) *Dashboard {
	d := &Dashboard{
		HealthyServersPanel: LineChart{
			Labels: []string{},
			Datasets: []Dataset{
				{
					Label:       "Healthy Servers",
					Data:        []int{},
					Tension:     0.1,
					BorderColor: "rgb(75, 192, 192)",
					Fill:        false,
				},
			},
		},
	}

	for _, im := range ims {
		d.HealthyServersPanel.Labels = append(d.HealthyServersPanel.Labels, fmt.Sprintf("%d", im.Timestamp.UnixMilli()))
		d.HealthyServersPanel.Datasets[0].Data = append(d.HealthyServersPanel.Datasets[0].Data, im.HealthyServersCount)
	}

	return d
}
