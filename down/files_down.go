package down

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func Down(url string, dst string, cor, timeout, size int, cache, verify bool, hash string) error {
	var length int
	var queue, redo, finish chan int
	if verify {
		file, err := os.Open(url)
		if err != nil {
			log.Infof("os.Open(url) err:%+v", err)
			return err
		}
		if hash == "sha1" {
			h := sha1.New()
			io.Copy(h, file)
			r := h.Sum(nil)
			log.Infof("sha1 of file: %x\n", r)
		} else if hash == "md5" {
			h := md5.New()
			io.Copy(h, file)
			r := h.Sum(nil)
			log.Infof("sha1 of file: %x\n", r)
		}
		return nil
	}

	if dst == "" {
		_, dst = filepath.Split(url)
	}

	startTime := time.Now()

	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {

		return err
	}
	response.Body.Close()
	num := response.Header.Get("Content-Length")
	length, _ = strconv.Atoi(num)
	log.Infoln("Conetnt-Length", length)
	ranges := response.Header.Get("Accept-Ranges")
	log.Infoln("response.Header.Get(\"Accept-Ranges\"):", ranges)

	if size <= 0 {
		size = int(math.Ceil(float64(length) / float64(cor)))
	}
	fragment := int(math.Ceil(float64(length) / float64(size)))
	fragmentPtr := &fragment
	log.Infof("fragment=%d, size=%d", fragment, size)
	queue = make(chan int, cor)
	redo = make(chan int, int(math.Floor(float64(cor)/2)))
	go func() {
		for i := 0; i < *fragmentPtr; i++ {
			log.Infof("queue <- %d", i)
			queue <- i
		}
		//redo是如果某块下载失败了，重新投递到queue，进而重新下载
		for {
			j := <-redo
			log.Infof("redo: queue <- %d", j)
			queue <- j
		}
	}()
	finish = make(chan int, cor)
	for j := 0; j < cor; j++ {
		go Do(request, fragmentPtr, j, size, length, timeout, dst, cache, queue, redo, finish)
	}

	//finish的目的：等分块fragment都完成了，主进程接着往下执行。如果没有这个，那么主进程不会等子协程结束就会提前退出
	for k := 0; k < fragment; k++ {
		_ = <-finish
		//log.Infof("[%s][%d]Finished\n", "-", i)
	}
	log.Infoln("Start to combine files...")

	file, err := os.Create(dst)
	if err != nil {
		log.Infoln(err)
		return err
	}
	defer file.Close()
	var offset int64 = 0

	//分块下载的多个文件，最后合并组装成一个
	for x := 0; x < fragment; x++ {
		filename := fmt.Sprintf("%s_%d", dst, x)
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Infoln(err)
			continue
		}
		file.WriteAt(buf, offset)
		offset += int64(len(buf))
		os.Remove(filename)
	}
	log.Infoln("Written to ", dst)
	//hash
	if hash == "sha1" {
		h := sha1.New()
		io.Copy(h, file)
		r := h.Sum(nil)
		log.Infof("sha1 of file: %x\n", r)
	} else if hash == "md5" {
		h := md5.New()
		io.Copy(h, file)
		r := h.Sum(nil)
		log.Infof("sha1 of file: %x\n", r)
	}

	finishTime := time.Now()
	duration := finishTime.Sub(startTime).Seconds()
	log.Infof("Time:%f Speed:%f Kb/s\n", duration, float64(length)/duration/1024)
	return nil
}

func Do(request *http.Request, fragmentPtr *int, no, size, length, timeout int, dst string, cache bool, queue, redo, finish chan int) error {
	log.Infof("Do request")
	var req http.Request
	err := DeepCopy(&req, request)
	if err != nil {
		log.Infoln("ERROR|prepare request:", err)
		return err
	}
	for {
		cStartTime := time.Now()

		i := <-queue
		//log.Infof("[%d][%d]Start download\n",no, i)
		start := i * size
		var end int
		if i < *fragmentPtr-1 {
			end = start + size - 1
		} else {
			end = length - 1
		}

		filename := fmt.Sprintf("%s_%d", dst, i)
		if cache {
			filesize := int64(end - start + 1)
			file, err := os.Stat(filename)
			if err == nil && file.Size() == filesize {
				log.Infof("[%d][%d]Hint cached %s, size:%d\n", no, i, filename, filesize)
				finish <- i
				continue
			}
		}

		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		log.Infof("[%d][%d]Start download:%d-%d\n", no, i, start, end)
		cli := http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
		resp, err := cli.Do(&req)
		if err != nil {
			log.Errorf("[%d][%d]ERROR|do request:%s\n", no, i, err.Error())
			redo <- i
			continue
		}

		//log.Infof("[%d]Content-Length:%s\n", i, resp.Header.Get("Content-Length"))
		log.Infof("[%d][%d]Content-Range:%s\n", no, i, resp.Header.Get("Content-Range"))

		file, err := os.Create(filename)
		if err != nil {
			log.Infof("[%d][%d]ERROR|create file %s:%s\n", no, i, filename, err.Error())
			file.Close()
			resp.Body.Close()
			redo <- i
			continue
		}
		log.Infof("[%d][%d]Writing to file %s\n", no, i, filename)
		n, err := io.Copy(file, resp.Body)
		if err != nil {
			log.Infof("[%d][%d]ERROR|write to file %s:%s\n", no, i, filename, err.Error())
			file.Close()
			resp.Body.Close()
			redo <- i
			continue
		}
		cEndTime := time.Now()
		duration := cEndTime.Sub(cStartTime).Seconds()
		log.Infof("[%d][%d]Download successfully:%f Kb/s\n", no, i, float64(n)/duration/1024)

		file.Close()
		resp.Body.Close()

		finish <- i
	}
	return nil
}

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
