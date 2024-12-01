package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	proxyHost = ""
	proxyPort = 12345
)

func getProxyUrl() *url.URL {
	proxyUrl, _ := url.Parse(fmt.Sprintf("http://%s:%d", proxyHost, proxyPort))
	return proxyUrl
}

func NewClient() *http.Client {
	if jar, err := cookiejar.New(nil); err == nil {
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(getProxyUrl()),
			},
			Jar: jar,
		}
		return httpClient
	}

	return nil
}

func NewClientWithAuth(username, password string) *http.Client {
	httpClient := NewClient()
	httpClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(&url.URL{
			Scheme: "http",
			User:   url.UserPassword(username, password),
			Host:   fmt.Sprintf("%s:%d", proxyHost, proxyPort),
		}),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return httpClient
}
