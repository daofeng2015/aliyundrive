package aliyundrive

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Object = map[string]interface{}
type Array = []Object

type AliyunDrive struct {
	refreshToken    string
	token           string
	tokenExpireTime int64
	driveID         string
	httpClient      *http.Client
}

func NewWithRefreshToken(refreshToken string) (*AliyunDrive, error) {
	d := &AliyunDrive{
		refreshToken: refreshToken,
		httpClient:   newHttpClient(),
	}
	if _, err := d.getToken(context.Background()); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *AliyunDrive) getToken(ctx context.Context) (string, error) {
	now := time.Now().Unix()
	if len(d.token) > 0 && now > d.tokenExpireTime {
		return d.token, nil
	}

	body := strings.NewReader(`{"refresh_token":"` + d.refreshToken + `"}`)
	request, err := http.NewRequestWithContext(ctx, "POST", ApiRefreshToken, body)
	if err != nil {
		return "", err
	}
	setCommonRequestHeader(request.Header)
	setJSONRequestHeader(request.Header)
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return "", err
	}

	ret := &struct {
		AccessToken    string `json:"access_token"`
		ExpiresIn      int64  `json:"expires_in"`
		DefaultDriveID string `json:"default_drive_id"`
		RefreshToken   string `json:"refresh_token"`
	}{}

	if err := json.Unmarshal(resp, ret); err != nil {
		return "", err
	}

	if len(ret.AccessToken) == 0 || len(ret.DefaultDriveID) == 0 || len(ret.RefreshToken) == 0 {
		return "", ErrRefreshToken
	}

	d.token = ret.AccessToken
	d.tokenExpireTime = now + ret.ExpiresIn
	d.driveID = ret.DefaultDriveID
	d.refreshToken = ret.RefreshToken
	return ret.AccessToken, nil
}
