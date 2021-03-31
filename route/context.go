package route

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	xhttp "github.com/HiData-xyz/hit/http"
	log "github.com/HiData-xyz/hit/log"

	jwt "github.com/dgrijalva/jwt-go"
)

var TokenHeader = "Authorization"

// NewContext 返回上下文实例
func NewContext(w http.ResponseWriter, r *http.Request, route *Route) *Context {
	return &Context{
		ctx:   context.Background(),
		w:     w,
		r:     r,
		val:   make(map[string]interface{}),
		Route: route,
	}
}

// Context 上下文环境
type Context struct {
	ctx   context.Context
	Route *Route

	w http.ResponseWriter
	r *http.Request

	val map[string]interface{}

	hooks  []*http.Request // 回调请求
	stoped int32
}

// Hook 回调函数
type Hook func(req *http.Request) error

func (ctx *Context) Context() context.Context {
	return ctx.ctx
}

func (ctx *Context) SetToken(header string, base string, v interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, v.(jwt.Claims))
	sign, err := token.SignedString([]byte(base))
	if err != nil {
		return "", err
	}
	ctx.w.Header().Add(header, sign)
	return sign, nil
}

// HTTPHook 生成HTTP回调
func (ctx *Context) HTTPHook(url string, body interface{}) (h Hook, err error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Error("序列化数据失败", "url", url, "err", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		log.Error(ErrNewHTTPRequestFail.Error(), "url", url, "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	return func(req *http.Request) error {
		err := xhttp.Do(req, 3, nil)
		if err != nil {
			log.Error(ErrHookFailed.Error(), "url", req.URL.String(), "err", err)
			return err
		}
		return nil
	}, nil
}

// SetHook 设置回调
func (ctx *Context) SetHook(url string, body interface{}) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Error("序列化数据失败", "url", url, "err", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		log.Error(ErrNewHTTPRequestFail.Error(), "url", url, "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	ctx.hooks = append(ctx.hooks, req)
}

// SetValue 设置值, 非并发安全
func (ctx *Context) SetValue(key string, val interface{}) {
	ctx.val[key] = val
}

// GetValue 获取设置的值, 非并发安全
func (ctx *Context) GetValue(key string) interface{} {
	return ctx.val[key]
}

// GetRequest 获取一份原始请求的深拷贝
func (ctx *Context) GetRequest() *http.Request {
	return ctx.r
}

// GetBody 获取请求body
func (ctx *Context) GetBody(a interface{}) (err error) {
	data, err := ctx.GetBodyBytes()
	if err != nil {
		return
	}
	return json.Unmarshal(data, a)
}

// GetBodyBytes 获取请求body
func (ctx *Context) GetBodyBytes() (data []byte, err error) {
	data, err = ioutil.ReadAll(ctx.r.Body)
	if err != nil {
		return nil, fmt.Errorf("%w\n%s", ErrReadRequestBodyFail, err.Error())
	}
	defer ctx.r.Body.Close()
	return
}

// GetString 获取URL携带的参数值
func (ctx *Context) GetString(key string) (val string) {
	return ctx.r.FormValue(key)
}

func (ctx *Context) GetInt(key string) (val int) {
	_v := ctx.r.FormValue(key)
	val, _ = strconv.Atoi(_v)
	return
}

func (ctx *Context) GetBool(key string) (val bool) {
	_v := ctx.r.FormValue(key)
	val, _ = strconv.ParseBool(_v)
	return
}

// Reset 重置状态和变量初始化
func (ctx *Context) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.w = w
	ctx.r = r
}

// JSON 返回 JSON 格式数据
func (ctx *Context) JSON(a interface{}) (err error) {
	if a == nil {
		return
	}

	var bytes []byte
	if b, ok := a.([]byte); ok {
		bytes = b
	} else {
		b, err := json.Marshal(a)
		if err != nil {
			return err
		}
		bytes = b
	}

	err = ctx.write(http.StatusOK, bytes)
	if err != nil {
		return
	}
	return
}

// EJSON 指定一个HTTP返回码, 通常是一个错误码
func (ctx *Context) EJSON(ecode int, a interface{}) (err error) {
	bytes, err := json.Marshal(a)
	if err != nil {
		return
	}
	err = ctx.write(ecode, bytes)
	if err != nil {
		return
	}
	return nil
}

func (ctx *Context) write(code int, b []byte) (err error) {
	defer ctx.Stop()

	ctx.w.Header().Set("Content-Type", "application/json")
	ctx.w.WriteHeader(code)
	_, err = ctx.w.Write(b)
	if err != nil {
		return
	}
	return
}

// Finish  公共处理
func (ctx *Context) Finish() {
	// 执行回调
	for _, val := range ctx.hooks {
		go func(req *http.Request) {
			err := xhttp.Do(req, 3, nil)
			if err != nil {
				log.Error(ErrHookFailed.Error(), "url", req.URL.String(), "err", err)
				return
			}
			log.Info("回调成功", "url", req.URL.String())
		}(val)
	}

	return
}

// Stop 停止执行方法链
func (ctx *Context) Stop() {
	atomic.CompareAndSwapInt32(&ctx.stoped, 0, 1)
}

// IsStop 是否停止
func (ctx *Context) IsStop() bool {
	return !atomic.CompareAndSwapInt32(&ctx.stoped, 0, 0)
}

// New 启动新的服务
func (ctx *Context) New(addr string, h http.Handler) {
	ctx.Route.AddServer(addr, h)
}

// ParseFiles 解析表单上传的文件
func (ctx *Context) ParseFiles(dir string) (files []*os.File, err error) {
	reader, err := ctx.r.MultipartReader()
	if err != nil {
		ctx.EJSON(http.StatusInternalServerError, err.Error())
		return
	}

	for {
		part, err := reader.NextPart()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err == io.EOF {
			break
		}
		fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
		if part.FileName() != "" {
			// part.FileName()得到的文件名以 "\\" 分割
			fileName := strings.ReplaceAll(part.FileName(), "\\", "/")
			fileName = filepath.Base(fileName)
			fileName = filepath.Join(dir, fileName)
			dst, err := os.Create(fileName)
			if err != nil {
				log.Error("创建文件失败", "err", err.Error(), "path", fileName)
				return nil, err
			}
			defer dst.Close()
			io.Copy(dst, part)

			files = append(files, dst)
		}
	}
	return
}
