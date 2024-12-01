package proxy

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	proxyHost = ""
	proxyPort = 12345
)

func NewClient() *http.Client {
	if proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%d", proxyHost, proxyPort)); err == nil {
		if jar, err := cookiejar.New(nil); err == nil {
			httpClient := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Jar: jar,
			}
			return httpClient
		}
	}

	return nil
}
