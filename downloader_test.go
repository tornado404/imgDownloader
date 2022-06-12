package imgDownloader

import (
	"path/filepath"
	"testing"
)

func TestDownloader(t *testing.T) {
	url := "xxx.jpg"
	dst := "/tmp"
	d := New()
	_, filename := filepath.Split(url)
	d.Download(url, filepath.Join(dst, filename))
}
