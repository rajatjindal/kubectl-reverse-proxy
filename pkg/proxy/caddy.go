package proxy

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
)

//go:embed caddyfile.tmpl
var caddyConfigTmpl string

type caddylb struct {
	listenPort string
	adminPort  string
	stopCh     <-chan struct{}
	reloadCh   chan map[string]string
	httpclient *http.Client
}

func NewCaddyReverseProxy(listenPort string, adminPort string, stopCh <-chan struct{}) *caddylb {
	return &caddylb{
		httpclient: &http.Client{
			Timeout: 2 * time.Second,
		},
		listenPort: listenPort,
		adminPort:  adminPort,
		stopCh:     stopCh,
		reloadCh:   make(chan map[string]string),
	}
}

func (c *caddylb) Start() {
	c.run()
}

func (c *caddylb) Reload(portmap map[string]string) {
	c.reloadCh <- portmap
}

func (c *caddylb) Stop() {
	c.stop()
}

func (c *caddylb) regenAndReload(portmap map[string]string) {
	config, err := c.regenConfig(portmap)
	if err != nil {
		fmt.Printf("failed to regen caddy config. %v\n", err)
		return
	}

	err = c.reload(config)
	if err != nil {
		fmt.Printf("failed to reload caddy config. %v\n", err)
		return
	}
}

func (c *caddylb) stop() error {
	req, err := http.NewRequest(http.MethodPost, c.getAdminEndpointUrl("/stop"), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *caddylb) reload(caddyconfig string) error {
	req, err := http.NewRequest(http.MethodPost, c.getAdminEndpointUrl("/load"), strings.NewReader(caddyconfig))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *caddylb) regenConfig(portMap map[string]string) (string, error) {
	ports := []string{}
	for _, value := range portMap {
		ports = append(ports, value)
	}

	config := struct {
		Ports        []string
		ListenOnPort string
		AdminPort    string
	}{
		Ports:        ports,
		ListenOnPort: c.listenPort,
		AdminPort:    c.adminPort,
	}

	tmpl, err := template.New("backendconfig").Parse(caddyConfigTmpl)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, config)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (c *caddylb) run() {
	caddy.Run(&caddy.Config{
		Admin: &caddy.AdminConfig{
			Listen: c.caddyAdminSvc(),
		},
	})

	for {
		select {
		case <-c.stopCh:
			fmt.Println("stopping caddy server")
			if err := c.stop(); err != nil {
				fmt.Println(err)
			}
		case portmap := <-c.reloadCh:
			c.regenAndReload(portmap)
		}
	}
}

func (c *caddylb) caddyAdminSvc() string {
	return fmt.Sprintf("localhost:%s", c.adminPort)
}

func (c *caddylb) getAdminEndpointUrl(path string) string {
	return fmt.Sprintf("http://%s%s", c.caddyAdminSvc(), path)
}
