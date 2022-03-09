package dns

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"snet/proxy"
	"snet/proxy/socks5"
	"time"
)

var client *http.Client

func init() {
	p, _ := proxy.Get("socks5")
	s := p.(*socks5.Server)
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(fmt.Sprintf("socks5://%s:%d", s.Host, s.Port))
			},
		},
	}
}

func cndoh(q []byte) (a []byte, err error) {
	return queryDOH(http.DefaultClient, "223.5.5.5", q)
}
func fqdoh(q []byte) (a []byte, err error) {
	return queryDOH(client, "8.8.4.4", q)
}
func queryDOH(client *http.Client, host string, q []byte) (a []byte, err error) {
	dohh := fmt.Sprintf("https://%s/dns-query", host)
	body := bytes.NewReader(q)
	timeout, cancel := context.WithTimeout(context.TODO(), dnsTimeout*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(timeout, "POST", dohh, body)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/dns-message")
	req.Header.Add("Content-Type", "application/dns-message")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTPS server returned with non-OK code %d", resp.StatusCode)
		return
	}
	return io.ReadAll(resp.Body)
}
