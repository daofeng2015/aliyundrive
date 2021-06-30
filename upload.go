package aliyundrive

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

type fileProof struct {
	DriveID       string      `json:"drive_id"`
	PartInfoList  []*partInfo `json:"part_info_list"`
	ParentFileID  string      `json:"parent_file_id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	CheckNameMode string      `json:"check_name_mode"`
	Size          int64       `json:"size"`
	PreHash       string      `json:"pre_hash"`
}

type partInfo struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url,omitempty"`
}

type createProofResponse struct {
	UploadID     string      `json:"upload_id"`
	FileID       string      `json:"file_id"`
	PartInfoList []*partInfo `json:"part_info_list"`
}

const (
	MaxPartSize = 1024 * 1024 * 1024 // 10M
)

func (d *Drive) UploadFromLocalFile(parentID string, p string) (*Item, error) {
	ctx := context.Background()
	return d.UploadFromLocalFileWithContext(ctx, parentID, p)
}

func (d *Drive) UploadFromLocalFileWithContext(ctx context.Context, parentID string, p string) (*Item, error) {
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return d.UploadWithContext(ctx, parentID, file)
}

func (d *Drive) Upload(parentID string, f fs.File) (*Item, error) {
	ctx := context.Background()
	return d.UploadWithContext(ctx, parentID, f)
}

func (d *Drive) UploadWithContext(ctx context.Context, parentID string, f fs.File) (*Item, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, ErrFileInvalid
	}

	name, size := path.Base(fileInfo.Name()), fileInfo.Size()
	proof := &fileProof{
		DriveID:       d.driveID,
		PartInfoList:  makePartInfoList(size),
		ParentFileID:  parentID,
		Name:          name,
		Type:          "file",
		CheckNameMode: "auto_rename",
		Size:          size,
		PreHash:       "",
	}
	proofResp, err := d.createFileWithProof(ctx, proof)
	if err != nil {
		return nil, err
	}

	for _, part := range proofResp.PartInfoList {
		partReader := io.LimitReader(f, MaxPartSize)
		err := d.uploadPart(ctx, part.UploadURL, partReader)
		if err != nil {
			return nil, err
		}
	}

	return d.completeUpload(ctx, proofResp)
}

func makePartInfoList(size int64) []*partInfo {
	partInfoNum := 0
	if size%MaxPartSize > 0 {
		partInfoNum++
	}
	partInfoNum += int(size / MaxPartSize)
	list := make([]*partInfo, partInfoNum)
	for i := 0; i < partInfoNum; i++ {
		list[i] = &partInfo{
			PartNumber: i + 1,
		}
	}
	return list
}

func (d *Drive) createFileWithProof(ctx context.Context, p *fileProof) (*createProofResponse, error) {
	body, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", apiCreateFileWithProof, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if err := d.setRequestHeaderAuthWithContext(ctx, request.Header); err != nil {
		return nil, err
	}
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return nil, err
	}

	proofResp := &createProofResponse{}
	if err := json.Unmarshal(resp, proofResp); err != nil {
		return nil, err
	}

	if len(proofResp.FileID) == 0 || len(proofResp.UploadID) == 0 || len(proofResp.PartInfoList) == 0 {
		return nil, ErrCreateFileWithProof
	}

	for _, part := range proofResp.PartInfoList {
		if len(part.UploadURL) == 0 {
			return nil, ErrCreateFileWithProof
		}
	}

	return proofResp, nil
}

func (d *Drive) uploadPart(ctx context.Context, api string, p io.Reader) error {
	request, err := http.NewRequestWithContext(ctx, "PUT", api, p)
	if err != nil {
		return err
	}
	resp, err := d.httpClient.Do(request)
	if err != nil {
		return err
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return ErrUploadPart
}

func (d *Drive) completeUpload(ctx context.Context, pr *createProofResponse) (*Item, error) {
	body, err := json.Marshal(Object{
		"drive_id":  d.driveID,
		"upload_id": pr.UploadID,
		"file_id":   pr.FileID,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", apiCompleteUpload, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if err := d.setRequestHeaderAuthWithContext(ctx, request.Header); err != nil {
		return nil, err
	}
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return nil, err
	}

	uploadResp := &Item{}
	if err := json.Unmarshal(resp, uploadResp); err != nil {
		return nil, err
	}
	return uploadResp, nil
}
