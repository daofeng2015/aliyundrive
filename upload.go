package aliyundriver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type FileInfo struct {
	Name    string
	Size    int
	PreHash string
}

type partInfo struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url,omitempty"`
}

type fileProof struct {
	DriveID       string      `json:"drive_id"`
	PartInfoList  []*partInfo `json:"part_info_list"`
	ParentFileID  string      `json:"parent_file_id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	CheckNameMode string      `json:"check_name_mode"`
	Size          int         `json:"size"`
	PreHash       string      `json:"pre_hash"`
}

type createProofResponse struct {
	UploadID     string      `json:"upload_id"`
	FileID       string      `json:"file_id"`
	PartInfoList []*partInfo `json:"part_info_list"`
}

type UploadResponse struct {
	FileID          string `json:"file_id"`
	Name            string `json:"name"`
	ContentType     string `json:"content_type"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	FileExtension   string `json:"file_extension"`
	Hidden          bool   `json:"hidden"`
	Size            int    `json:"size"`
	Starred         bool   `json:"starred"`
	Status          string `json:"status"`
	UploadID        string `json:"upload_id"`
	ParentFileID    string `json:"parent_file_id"`
	CRC64Hash       string `json:"crc64_hash"`
	ContentHash     string `json:"content_hash"`
	ContentHashName string `json:"content_hash_name"`
	Category        string `json:"category"`
	EncryptMode     string `json:"encrypt_mode"`
	Location        string `json:"location"`
}

const (
	MaxPartSize = 1024 * 1024 * 1024 // 10M
	FakeUA      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"

	ApiCreateFileWithProof = "https://api.aliyundrive.com/v2/file/create_with_proof"
	ApiCompleteUpload      = "https://api.aliyundrive.com/v2/file/complete"
)

var (
	ErrCreateFileWithProof = errors.New("ap create_with_proof failed")
	ErrUploadPart          = errors.New("upload part file failed")
)

func (d *AliyunDriver) Upload(ctx context.Context, parentID string, info *FileInfo, f io.Reader) (*UploadResponse, error) {
	proof := d.newFileProof(parentID, info)
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

	return d.complieteUpload(ctx, proofResp)
}

func (d *AliyunDriver) newFileProof(parentID string, info *FileInfo) *fileProof {
	p := &fileProof{
		DriveID:       d.config.DriveID,
		PartInfoList:  makePartInfoList(info.Size),
		ParentFileID:  parentID,
		Name:          info.Name,
		Type:          "file",
		CheckNameMode: "auto_rename",
		Size:          info.Size,
		PreHash:       info.PreHash,
	}
	return p
}

func (d *AliyunDriver) createFileWithProof(ctx context.Context, p *fileProof) (*createProofResponse, error) {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(p); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", ApiCreateFileWithProof, body)
	if err != nil {
		return nil, err
	}
	setRequestHeader(request.Header)
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("accept", "application/json, text/plain, */*")
	request.Header.Set("authorization", "Bearer "+d.config.Token)
	resp, err := d.httpClient.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	proofResp := &createProofResponse{}
	if err := json.NewDecoder(resp.Body).Decode(proofResp); err != nil {
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

func (d *AliyunDriver) uploadPart(ctx context.Context, api string, p io.Reader) error {
	request, err := http.NewRequestWithContext(ctx, "PUT", api, p)
	if err != nil {
		return err
	}

	resp, err := d.httpClient.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}

	io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return ErrUploadPart
}

func (d *AliyunDriver) complieteUpload(ctx context.Context, pr *createProofResponse) (*UploadResponse, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(map[string]string{
		"drive_id":  d.config.DriveID,
		"upload_id": pr.UploadID,
		"file_id":   pr.FileID,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", ApiCompleteUpload, body)
	if err != nil {
		return nil, err
	}
	setRequestHeader(request.Header)
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("accept", "application/json, text/plain, */*")
	request.Header.Set("authorization", "Bearer "+d.config.Token)
	resp, err := d.httpClient.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	uploadResp := &UploadResponse{}
	if err := json.NewDecoder(resp.Body).Decode(uploadResp); err != nil {
		return nil, err
	}
	return uploadResp, nil
}

func makePartInfoList(size int) []*partInfo {
	partInfoNum := 0
	if size%MaxPartSize > 0 {
		partInfoNum++
	}
	partInfoNum += size / MaxPartSize
	list := make([]*partInfo, partInfoNum)
	for i := 0; i < partInfoNum; i++ {
		list[i] = &partInfo{
			PartNumber: i + 1,
		}
	}
	return list
}

func setRequestHeader(header http.Header) {
	header.Set("origin", "https://www.aliyundrive.com")
	header.Set("referer", "https://www.aliyundrive.com/")
	header.Set("pragma", "no-cache")
	header.Set("dnt", "1")
	header.Set("cache-control", "no-cache")
	header.Set("user-agent", FakeUA)
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7,zh-TW;q=0.6")
}
