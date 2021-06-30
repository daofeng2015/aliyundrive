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

type Drive struct {
	refreshToken    string
	token           string
	tokenExpireTime int64
	driveID         string
	rootID          string
	httpClient      *http.Client
}

type Item struct {
	DriveID         string    `json:"drive_id"`
	DomainID        string    `json:"domain_id"`
	UploadID        string    `json:"upload_id"`
	FileID          string    `json:"file_id"`
	ParentFileID    string    `json:"parent_file_id"`
	Name            string    `json:"name"`
	FileExtension   string    `json:"file_extension"`
	Size            uint64    `json:"size"`
	Type            string    `json:"type"`
	ContentType     string    `json:"content_type"`
	MimeExtension   string    `json:"mime_extension"`
	MimeType        string    `json:"mime_type"`
	Category        string    `json:"category"`
	Hidden          bool      `json:"hidden"`
	Status          string    `json:"status"`
	Starred         bool      `json:"starred"`
	Trashed         bool      `json:"trashed"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	EncryptMode     string    `json:"encrypt_mode"`
	Location        string    `json:"location"`
	CRC64Hash       string    `json:"crc64_hash"`
	ContentHash     string    `json:"content_hash"`
	ContentHashName string    `json:"content_hash_name"`
	URL             string    `json:"url"`
	DownloadURL     string    `json:"download_url"`
	Thumbnail       string    `json:"thumbnail"`
	Labels          []string  `json:"labels"`
}

func NewWithRefreshToken(refreshToken string) (*Drive, error) {
	d := &Drive{
		refreshToken: refreshToken,
		rootID:       "root",
		httpClient:   newHttpClient(),
	}
	if _, err := d.getToken(context.Background()); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Drive) SetRootID(rootID string) {
	d.rootID = rootID
}

func (d *Drive) GetRootID() string {
	return d.rootID
}

func (d *Drive) getToken(ctx context.Context) (string, error) {
	now := time.Now().Unix()
	if len(d.token) > 0 && now > d.tokenExpireTime {
		return d.token, nil
	}

	body := strings.NewReader(`{"refresh_token":"` + d.refreshToken + `,"grant_type":"refresh_token""}`)
	request, err := http.NewRequestWithContext(ctx, "POST", apiRefreshToken, body)
	if err != nil {
		return "", err
	}
	setCommonRequestHeader(request.Header)
	setJSONRequestHeader(request.Header)
	resp,code,err := d.DoRequestBytes(request)
	if err != nil {
		return "", err
	}
	if code !=200{
		return "refresh_token err",err
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
