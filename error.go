package aliyundrive

import "errors"

var (
	ErrRefreshToken        = errors.New("refresh token failed")
	ErrFileInvalid         = errors.New("invalid file given")
	ErrCreateFileWithProof = errors.New("api create_with_proof failed")
	ErrUploadPart          = errors.New("upload part file failed")
	ErrGetDownloadURL      = errors.New("api get_download_url faied")
)
