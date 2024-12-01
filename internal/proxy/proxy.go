package proxy

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	proxyHost = ""
	proxyPort = 12345
)

type proxyAuth struct {
	username string
	password string
}

func (p *proxyAuth) basicAuth() string {
	auth := p.username + ":" + p.password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

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
	proxyAuth := &proxyAuth{username, password}
	httpClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(getProxyUrl()),
		ProxyConnectHeader: http.Header{
			"Proxy-Authorization": []string{proxyAuth.basicAuth()},
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return httpClient
}
