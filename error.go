package aliyundrive

import "errors"

var (
	ErrRefreshToken         = errors.New("refresh token failed")
	ErrFileInvalid          = errors.New("invalid file given")
	ErrCreateFileWithProof  = errors.New("api create_with_proof failed")
	ErrUploadPart           = errors.New("upload part file failed")
	ErrGetDownloadURL       = errors.New("api get_download_url faied")
	ErrUnexpectedStatusCode = errors.New("unexpected status code get")
	ErrOpenItemNotFile      = errors.New("item to open is not kind of file")
	ErrRemoveFailed         = errors.New("remove item failed")
	ErrBatchRequestFailed   = errors.New("batch request failed")
)
