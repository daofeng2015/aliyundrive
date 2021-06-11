package aliyundriver

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type AliyunDriver struct {
	config *AliyunDriverConfig

	httpClient *http.Client
}

type AliyunDriverConfig struct {
	DriveID string
	Token   string
}

func NewAliyunDriver(config *AliyunDriverConfig) *AliyunDriver {
	return &AliyunDriver{
		config:     config,
		httpClient: newHttpClient(),
	}
}

func newHttpClient() *http.Client {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c := &http.Client{
		Transport: t,
		Timeout:   time.Second * 3,
	}
	return c
}
