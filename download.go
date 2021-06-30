package aliyundrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func (d *AliyunDrive) DownloadLocalFile(fileID string, target string) (int64, error) {
	ctx := context.Background()
	return d.DownloadLocalFileWithContext(ctx, fileID, target)
}

func (d *AliyunDrive) DownloadLocalFileWithContext(ctx context.Context, fileID string, target string) (int64, error) {
	targetFile, err := os.Create(target)
	if err != nil {
		return 0, err
	}
	defer targetFile.Close()
	remoteReader, err := d.DownloadWithContext(ctx, fileID)
	if err != nil {
		return 0, err
	}
	defer remoteReader.Close()
	return io.Copy(targetFile, remoteReader)
}

func (d *AliyunDrive) Download(fileID string) (io.ReadCloser, error) {
	ctx := context.Background()
	return d.DownloadWithContext(ctx, fileID)
}

func (d *AliyunDrive) DownloadWithContext(ctx context.Context, fileID string) (io.ReadCloser, error) {
	downloadURL, err := d.GetDownloadURLWithContext(ctx, fileID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	setCommonRequestHeader(request.Header)
	resp, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, ErrUnexpectedStatusCode
	}
	return resp.Body, nil
}

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
