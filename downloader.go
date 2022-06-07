package main

import (
	"github.com/tornado404/imgDownloader/down"
	"path/filepath"
)

type Interface interface {
}

type FilesDownloader struct {
	CoroutineNum     int
	ChunkSize        int
	Timeout          int
	Dst              string // file name
	Hash             string // "sha1 or md5 to verify the file"
	IsVerifyRequired bool   // "verify file, not download"
	Cache            bool   // jump if cache exist, only verify the size
}

func New() *FilesDownloader {
	return &FilesDownloader{
		CoroutineNum:     10,
		ChunkSize:        1000000,
		Timeout:          0,
		Dst:              "",
		Hash:             "sha1",
		IsVerifyRequired: false,
		Cache:            false,
	}
}

func (f *FilesDownloader) DownloadBatch(urls map[string]string, dstPath string) {
	for fName, url := range urls {
		filepath.Join(dstPath, fName)
		f.Download(url, dstPath)
	}
}

func (f *FilesDownloader) Download(url string, dst string) error {
	return down.Down(url, dst, f.CoroutineNum, f.Timeout, f.ChunkSize, f.Cache, f.IsVerifyRequired, f.Hash)
}
