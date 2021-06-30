package aliyundrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type LSLimit uint
type OrderBy string
type OrderDirection string

const (
	OrderByName      OrderBy = "name"
	OrderByCreatedAt OrderBy = "created_at"
	OrderByUpdatedAt OrderBy = "updated_at"
	OrderBySize      OrderBy = "size"
	orderByDefault           = OrderByName

	OrderDirectionDesc    OrderDirection = "DESC"
	OrderDirectionAsc     OrderDirection = "ASC"
	orderDirectionDefault                = OrderDirectionDesc

	LSLimitUnlimited LSLimit = 0
	lSLimitDefault           = LSLimitUnlimited
)

type lsOption struct {
	limit          LSLimit
	orderBy        OrderBy
	orderDirection OrderDirection
}

type lsOptionSetter = func(opt *lsOption)

type LSResponse struct {
	Items      []*Item `json:"items"`
	NextMarker string  `json:"next_marker"`
}

func OptionLimit(l LSLimit) lsOptionSetter {
	return func(opt *lsOption) {
		opt.limit = l
	}
}

func OptionOrderBy(o OrderBy) lsOptionSetter {
	return func(opt *lsOption) {
		opt.orderBy = o
	}
}

func OptionOrderDirection(d OrderDirection) lsOptionSetter {
	return func(opt *lsOption) {
		opt.orderDirection = d
	}
}

func (d *Drive) ListItems(parentID string, marker string, opts ...lsOptionSetter) (*LSResponse, error) {
	if parentID == "" {
		parentID = d.rootID
	}
	opt := &lsOption{
		limit:          lSLimitDefault,
		orderBy:        orderByDefault,
		orderDirection: orderDirectionDefault,
	}
	for _, optSetter := range opts {
		optSetter(opt)
	}

	params := Object{
		"drive_id":                d.driveID,
		"order_by":                opt.orderBy,
		"order_direction":         opt.orderDirection,
		"parent_file_id":          parentID,
		"fields":                  "*",
		"image_thumbnail_process": "image/resize,w_400/format,jpeg",
		"image_url_process":       "image/resize,w_1920/format,jpeg",
		"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
		"url_expire_sec":          1600,
	}

	if opt.limit == LSLimitUnlimited {
		params["all"] = true
		params["limit"] = 9999999999
	} else {
		params["all"] = false
		params["limit"] = opt.limit
	}

	if len(marker) > 0 {
		params["marker"] = marker
	}

	retv := &LSResponse{
		Items:      make([]*Item, 0),
		NextMarker: "",
	}

	for {
		if len(retv.NextMarker) > 0 {
			params["marker"] = retv.NextMarker
		}
		body, err := json.Marshal(params)
		if err != nil {
			return retv, err
		}
		request, err := http.NewRequest("POST", apiList, bytes.NewReader(body))
		if err != nil {
			return retv, err
		}

		if err := d.setRequestHeaderAuth(request.Header); err != nil {
			return retv, err
		}

		resp, err := d.DoRequestBytes(request)
		if err != nil {
			return retv, err
		}

		lsresp := &LSResponse{}
		if err := json.Unmarshal(resp, lsresp); err != nil {
			return retv, err
		}

		if opt.limit != LSLimitUnlimited {
			return lsresp, nil
		}

		retv.Items = append(retv.Items, lsresp.Items...)
		retv.NextMarker = lsresp.NextMarker
		if len(lsresp.NextMarker) == 0 {
			break
		}
	}
	return retv, nil
}

func (d *Drive) MkDir(parentFileID string, name string) (*Item, error) {
	params := Object{
		"check_name_mode": "auto_rename",
		"type":            "folder",
		"drive_id":        d.driveID,
		"name":            name,
		"parent_file_id":  parentFileID,
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", apiCreateWithFolder, bytes.NewReader(body))
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

	ret := &struct {
		DomainID     string `json:"domain_id"`
		DriveID      string `json:"drive_id"`
		FileID       string `json:"file_id"`
		ParentFileID string `json:"parent_file_id"`
		Name         string `json:"file_name"`
		Type         string `json:"type"`
		EncryptMode  string `json:"encrypt_mode"`
	}{}

	if err := json.Unmarshal(resp, ret); err != nil {
		return nil, err
	}

	item := &Item{
		DomainID:     ret.DomainID,
		DriveID:      ret.DriveID,
		FileID:       ret.FileID,
		ParentFileID: ret.ParentFileID,
		Name:         ret.Name,
		Type:         ret.Type,
		EncryptMode:  ret.EncryptMode,
	}
	return item, nil
}

func (d *Drive) Remove(fileID string, force bool) error {
	body := fmt.Sprintf(`{"drive_id":"%v","file_id":"%v"}`, d.driveID, fileID)
	api := apiTrash
	if force {
		api = apiDelete
	}
	request, err := http.NewRequest("POST", api, strings.NewReader(body))
	if err != nil {
		return err
	}
	if err := d.setRequestHeaderAuth(request.Header); err != nil {
		return err
	}
	resp, err := d.DoRequestBytes(request)
	if err != nil {
		return err
	}
	ret := &struct {
		TaskID   string `json:"async_task_id"`
		DomainID string `json:"domain_id"`
		DriveID  string `json:"drive_id"`
		FileID   string `json:"file_id"`
	}{}

	if err := json.Unmarshal(resp, ret); err != nil {
		return err
	}

	if len(ret.TaskID) == 0 || len(ret.FileID) == 0 {
		return ErrRemoveFailed
	}

	return nil
}

func (d *Drive) BatchRemove(fileIDList []string, force bool) ([]error, error) {
	api := "/recyclebin/trash"
	if force {
		api = "/file/delete"
	}
	requests := Array{}
	for _, fileID := range fileIDList {
		requests = append(requests, Object{
			"url":    api,
			"method": "POST",
			"id":     fileID,
			"header": Object{"Content-Type": "application/json"},
			"body":   Object{"drive_id": d.driveID, "file_id": fileID},
		})
	}

	params := Object{
		"requests": requests,
		"resource": "file",
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", apiBatch, bytes.NewReader(body))
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

	ret := &struct {
		Responses []struct {
			ID     string `json:"id"`
			Status int    `json:"status"`
			Body   struct {
				TaskID   string `json:"async_task_id"`
				DomainID string `json:"domain_id"`
				DriveID  string `json:"drive_id"`
				FileID   string `json:"file_id"`
			} `json:"body"`
		} `json:"responses"`
	}{}

	if err := json.Unmarshal(resp, ret); err != nil {
		return nil, err
	}

	if len(ret.Responses) != len(fileIDList) {
		return nil, ErrBatchRequestFailed
	}

	errList := make([]error, 0)
	for _, response := range ret.Responses {
		if len(response.Body.TaskID) == 0 || len(response.Body.FileID) == 0 {
			errList = append(errList, ErrRemoveFailed)
		} else {
			errList = append(errList, nil)
		}
	}
	return errList, nil
}
