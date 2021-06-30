package aliyundrive

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	ApiRefreshToken        = "https://websv.aliyundrive.com/token/refresh"
	ApiCreateFileWithProof = "https://api.aliyundrive.com/v2/file/create_with_proof"
	ApiCompleteUpload      = "https://api.aliyundrive.com/v2/file/complete"
	ApiGetDownloadURL      = "https://api.aliyundrive.com/v2/file/get_download_url"

	FakeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"
)

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

func setCommonRequestHeader(header http.Header) {
	header.Set("origin", "https://www.aliyundrive.com")
	header.Set("referer", "https://www.aliyundrive.com/")
	header.Set("pragma", "no-cache")
	header.Set("dnt", "1")
	header.Set("cache-control", "no-cache")
	header.Set("user-agent", FakeUA)
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7,zh-TW;q=0.6")
}

func setJSONRequestHeader(header http.Header) {
	header.Set("content-type", "application/json;charset=UTF-8")
	header.Set("accept", "application/json, text/plain, */*")
}

func (d *AliyunDrive) DoRequestBytes(request *http.Request) ([]byte, error) {
	resp, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return data, err
}
