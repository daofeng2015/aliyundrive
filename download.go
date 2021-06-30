package aliyundrive

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func (d *Drive) DownloadToLocalFile(fileID string, target string) (int64, error) {
	ctx := context.Background()
	return d.DownloadToLocalFileWithContext(ctx, fileID, target)
}

func (d *Drive) DownloadToLocalFileWithContext(ctx context.Context, fileID string, target string) (int64, error) {
	targetFile, err := os.Create(target)
	if err != nil {
		return 0, err
	}
	defer targetFile.Close()
	remoteReader, err := d.OpenItemFileWithContext(ctx, fileID)
	if err != nil {
		return 0, err
	}
	defer remoteReader.Close()
	return io.Copy(targetFile, remoteReader)
}

func (d *Drive) OpenItemFile(fileID string) (io.ReadCloser, error) {
	ctx := context.Background()
	return d.OpenItemFileWithContext(ctx, fileID)
}

func (d *Drive) OpenItemFileWithContext(ctx context.Context, fileID string) (io.ReadCloser, error) {
	item, err := d.GetItem(fileID)
	if err != nil {
		return nil, err
	}
	if item.Type != "file" {
		return nil, ErrOpenItemNotFile
	}
	request, err := http.NewRequest("GET", item.DownloadURL, nil)
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

func (d *Drive) GetItem(fileID string) (*Item, error) {
	params := Object{
		"drive_id":                d.driveID,
		"file_id":                 fileID,
		"image_thumbnail_process": "image/resize,w_400/format,jpeg",
		"image_url_process":       "image/resize,w_1920/format,jpeg",
		"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
		"url_expire_sec":          1600,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", apiFileGet, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if err := d.setRequestHeaderAuth(request.Header); err != nil {
		return nil, err
	}
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return nil, err
	}
	ret := &Item{}
	if err := json.Unmarshal(resp, ret); err != nil {
		return nil, err
	}
	return ret, nil
}
