package aliyundriver

import (
	"os"
	"testing"
)

func TestUpload(t *testing.T) {

	refreshToken := ``
	parentID := ""
	path := ""

	c, err := NewAliyunDriver(refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Upload(parentID, stat, f)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.FileID)

	u, err := c.GetDownloadURL(resp.FileID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(u)
}
