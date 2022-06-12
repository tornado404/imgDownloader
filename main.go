package imgDownloader

import (
	"flag"
	"fmt"
	_ "github.com/sirupsen/logrus"
	_ "github.com/tornado404/imgDownloader/down"
	"os"
	"path/filepath"
)

const VERSION = "1.0.3"

func main() {
	var dst string
	//flag.IntVar(&cor, "c", 1, "coroutine num")
	//flag.IntVar(&size, "s", 0, "chunk size")
	//flag.IntVar(&size, "t", 0, "timeout")
	flag.StringVar(&dst, "f", "", "file name")
	//flag.StringVar(&hash, "h", "sha1", "sha1 or md5 to verify the file")
	//flag.BoolVar(&verify, "v", false, "verify file, not download")
	//flag.BoolVar(&cache, "cache", false, "jump if cache exist, only verify the size")
	//flag.BoolVar(&version, "version", false, "show version")
	flag.Parse()

	var version bool
	url := os.Args[len(os.Args)-1]

	if version || url == "version" {
		fmt.Println("downloader version:", VERSION)
		return
	}
	d := New()
	_, filename := filepath.Split(url)
	d.Download(url, filepath.Join(dst, filename))
}
