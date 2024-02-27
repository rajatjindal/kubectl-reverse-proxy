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
			resp, err := s.client.Get(s.rvaddr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Println("unexpected code ", resp.StatusCode)
				continue
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			resp.Body.Close()

			err = s.parseOnce(raw)
			if err != nil {
				fmt.Println(err)
			}
		case <-s.stopCh:
			return nil
		}
	}
}

func (s *server) parseOnce(raw []byte) error {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	imData := processMetrics(mf,
		processHealthyServersCount,
	)

	s.Lock()
	defer s.Unlock()

	s.instanceMetrics = append(s.instanceMetrics, imData)

	return nil
}

type Dashboard struct {
	HealthyServersPanel LineChart `json:"HealthyServersPanel"`
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
