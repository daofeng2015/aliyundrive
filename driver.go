package aliyundriver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type AliyunDriver struct {
	RefreshToken string

	token           string
	tokenExpireTime int64
	driveID         string

	httpClient *http.Client
}

var (
	ErrRefreshToken   = errors.New("refresh token failed")
	ErrGetDownloadURL = errors.New("api get_download_url faied")
)

func NewAliyunDriver(refreshToken string) (*AliyunDriver, error) {
	d := &AliyunDriver{
		RefreshToken: refreshToken,
		httpClient:   newHttpClient(),
	}
	if _, err := d.getToken(context.Background()); err != nil {
		return nil, err
	}
	return d, nil
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

func (d *AliyunDriver) GetDownloadURL(fileID string) (string, error) {
	ctx := context.Background()
	return d.GetDownloadURLWithContext(ctx, fileID)
}

func (d *AliyunDriver) GetDownloadURLWithContext(ctx context.Context, fileID string) (string, error) {
	token, err := d.getToken(ctx)
	if err != nil {
		return "", err
	}
	params := fmt.Sprintf(`{"drive_id":"%v","file_id":"%v"}`, d.driveID, fileID)
	body := strings.NewReader(params)
	request, err := http.NewRequestWithContext(ctx, "POST", ApiGetDownloadURL, body)
	if err != nil {
		return "", err
	}
	setRequestHeader(request.Header)
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("accept", "application/json, text/plain, */*")
	request.Header.Set("authorization", "Bearer "+token)
	resp, err := d.httpClient.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return "", err
	}

	ret := &struct {
		URL string `json:"url"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return "", err
	}
	if len(ret.URL) == 0 {
		return "", ErrGetDownloadURL
	}
	return ret.URL, nil
}

func (d *AliyunDriver) getToken(ctx context.Context) (string, error) {
	now := time.Now().Unix()
	if len(d.token) > 0 && now > d.tokenExpireTime {
		return d.token, nil
	}

	body := strings.NewReader(`{"refresh_token":"` + d.RefreshToken + `"}`)
	request, err := http.NewRequestWithContext(ctx, "POST", ApiRefreshToken, body)
	if err != nil {
		return "", err
	}
	setRequestHeader(request.Header)
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("accept", "application/json, text/plain, */*")

	resp, err := d.httpClient.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	ret := &struct {
		AccessToken    string `json:"access_token"`
		ExpiresIn      int64  `json:"expires_in"`
		DefaultDriveID string `json:"default_drive_id"`
		RefreshToken   string `json:"refresh_token"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return "", err
	}

	if len(ret.AccessToken) == 0 || len(ret.DefaultDriveID) == 0 || len(ret.RefreshToken) == 0 {
		return "", ErrRefreshToken
	}

	d.token = ret.AccessToken
	d.tokenExpireTime = now + ret.ExpiresIn
	d.driveID = ret.DefaultDriveID
	d.RefreshToken = ret.RefreshToken

	return ret.AccessToken, nil
}
