package aliyundrive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (d *AliyunDrive) GetDownloadURL(fileID string) (string, error) {
	ctx := context.Background()
	return d.GetDownloadURLWithContext(ctx, fileID)
}

func (d *AliyunDrive) GetDownloadURLWithContext(ctx context.Context, fileID string) (string, error) {
	token, err := d.getToken(ctx)
	if err != nil {
		return "", err
	}
	params := fmt.Sprintf(`{"drive_id":"%v","file_id":"%v"}`, d.driveID, fileID)
	request, err := http.NewRequestWithContext(ctx,
		"POST", ApiGetDownloadURL, strings.NewReader(params))
	if err != nil {
		return "", err
	}
	setCommonRequestHeader(request.Header)
	setJSONRequestHeader(request.Header)
	request.Header.Set("Authorization", "Bearer "+token)
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return "", err
	}

	ret := &struct {
		URL string `json:"url"`
	}{}
	if err := json.Unmarshal(resp, ret); err != nil {
		return "", err
	}
	if len(ret.URL) == 0 {
		return "", ErrGetDownloadURL
	}
	return ret.URL, nil
}
