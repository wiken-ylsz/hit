package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	xhttp "net/http"
	"os"
	"time"

	"github.com/HiData-xyz/hit/log"
)

// Get 发送GET请求, 数据传输格式使用JSON
func Get(url string, res interface{}) (err error) {
	log.Info("发送GET请求", "URL:", url)
	return GetTimes(url, 3, res)
}

// Post 发送POST请求, 数据传输格式使用JSON
func Post(url string, req, res interface{}) (err error) {
	log.Info("发送POST请求", "URL:", url)
	return PostTimes(url, 3, req, res)
}

var defaultClient = xhttp.Client{
	Timeout: 10 * 60 * time.Second,
}

// GetTimes 发送GET请求, 数据传输格式使用JSON
// times: 请求失败后重试次数
func GetTimes(url string, times int, res interface{}) (err error) {
	httpReq, err := xhttp.NewRequest(xhttp.MethodGet, url, nil)
	if err != nil {
		return
	}
	err = Do(httpReq, times, res)
	if err != nil {
		return
	}

	return
}

// PostTimes 发送POST请求, 数据传输格式使用JSON
// times: 请求失败后重试次数
func PostTimes(url string, times int, req, res interface{}) (err error) {
	// 转化对象成json数据
	reqBody, err := json.Marshal(req)
	if err != nil {
		return
	}

	httpReq, err := xhttp.NewRequest(xhttp.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	err = Do(httpReq, times, res)
	if err != nil {
		return
	}

	return
}

// Do 发送请求
func Do(req *xhttp.Request, times int, res interface{}) error {
	for i := 0; i < times; i++ {
		// http请求
		response, err := defaultClient.Do(req)
		if err != nil {
			if i < times-1 {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return err
		}
		defer response.Body.Close()

		// 读取返回数据
		data, err := ioutil.ReadAll(response.Body)
		if err != nil && i >= times-1 {
			return err
		}

		if code := response.StatusCode; 299 < code || code < 200 {
			log.Info(fmt.Sprintf("code: %d", code))
			return errors.New(string(data))
		}

		if res == nil {
			return nil
		}

		log.Info("返回结果", "URL", req.URL.Path, "body", string(data))

		err = json.Unmarshal(data, res)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

// UploadFiles 上传文件
func UploadFiles(srcFile, filename, url string) (err error) {
	f, err := os.Open(srcFile)
	if err != nil {
		return
	}
	defer f.Close()

	var buff bytes.Buffer

	mulWriter := multipart.NewWriter(&buff)
	w, err := mulWriter.CreateFormFile(filename, filename)
	if err != nil {
		return
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return
	}
	mulWriter.Close()

	r, err := xhttp.NewRequest(xhttp.MethodPost, url, &buff)
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", mulWriter.FormDataContentType())
	err = Do(r, 3, nil)
	if err != nil {
		return
	}
	return
}
