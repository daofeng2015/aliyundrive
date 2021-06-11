package aliyundriver

import "net/http"

const (
	ApiRefreshToken        = "https://websv.aliyundrive.com/token/refresh"
	ApiCreateFileWithProof = "https://api.aliyundrive.com/v2/file/create_with_proof"
	ApiCompleteUpload      = "https://api.aliyundrive.com/v2/file/complete"
	ApiGetDownloadURL      = "https://api.aliyundrive.com/v2/file/get_download_url"

	FakeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"
)

func setRequestHeader(header http.Header) {
	header.Set("origin", "https://www.aliyundrive.com")
	header.Set("referer", "https://www.aliyundrive.com/")
	header.Set("pragma", "no-cache")
	header.Set("dnt", "1")
	header.Set("cache-control", "no-cache")
	header.Set("user-agent", FakeUA)
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7,zh-TW;q=0.6")
}
