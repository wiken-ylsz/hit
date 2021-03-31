package route

import (
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"

	log "github.com/HiData-xyz/hit/log"
)

// 常用错误
var (
	ErrMainServerClose    = errors.New("主服务已关闭")
	ErrHookFailed         = errors.New("回调失败")
	ErrNewHTTPRequestFail = errors.New("构建http请求失败")
)

// Handler 注册函数
type Handler func(ctx *Context)

// Router 一个简易的HTTP路由
type Router interface {
	// 启动HTTP服务
	Run(addr string) error
	// 注册路由, 覆盖已存在
	Post(path string, h ...Handler)
	Get(path string, h ...Handler)
	Group(path string, h ...Handler) *Group

	Error(err error) //

	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Handles 处理器
type Handles []Handler

// Route HTTP 路由器
type Route struct {
	ctx    context.Context
	cancel context.CancelFunc
	svr    *http.Server

	log log.Logger

	root   *tree
	paths  map[string]map[string]Handles // 路由
	middle Handles                       // 使用的中间件

	m        sync.Mutex
	parent   *Route            // 父级服务
	children map[string]*Route // 子服务

	wg sync.WaitGroup // 同步锁等待

	err     error // 处理链方法执行过程中产生的错误信息
	isPprof bool  // 是否开启性能监控
}

// AddServer 添加、运行子服务
func (r *Route) AddServer(addr string, h http.Handler) {
	var _r *Route
	// Route 内置注册性能监控路由
	if rou, ok := h.(*Route); ok {
		_r = rou
	} else {
		_r = New()
		_r.svr.Handler = h
	}

	_r.ctx, _r.cancel = context.WithCancel(r.ctx)
	r.m.Lock()
	defer r.m.Unlock()
	if _, ok := r.children[addr]; ok {
		return
	}
	r.children[addr] = _r

	go _r.Run(addr)
}

// SetPprof 设置是否开启性能监控
func (r *Route) SetPprof(b bool) {
	r.isPprof = b
}

// DeleteServer 停止、删除子服务
// TODO: 服务安全退出, 处理完已接受的请求
func (r *Route) DeleteServer(addr string) {
	r.m.Lock()
	defer r.m.Unlock()
	if _, ok := r.children[addr]; !ok {
		return
	}
	delete(r.children, addr)
}

// ServeHTTP 实现 HTTP.Server 接口
func (r *Route) ServeHTTP(w http.ResponseWriter, _r *http.Request) {
	ctx := NewContext(w, _r, r)
	ctx.Reset(w, _r)

	// 执行中间件
	for _, h := range r.middle {
		h(ctx)
		if ctx.IsStop() {
			return
		}
	}

	handles := r.Match(_r.URL.Path, _r.Method)
	if len(handles) == 0 {
		ctx.EJSON(404, "页面不存在")
		return
	}

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Info(fmt.Sprintf("%+v", err))
	// 		ctx.EJSON(http.StatusInternalServerError, err)
	// 	}
	// }()

	defer ctx.Finish()
	// 解析URL、表单参数
	_r.ParseForm()
	for _, h := range handles {
		h(ctx)
		if ctx.IsStop() {
			return
		}
	}

}

// Run 启动 HTTP 服务
func (r *Route) Run(addr string) error {
	select {
	case <-r.ctx.Done():
		return ErrMainServerClose
	default:
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		select {
		case <-r.ctx.Done():
			r.svr.Shutdown(context.Background())
			if r.parent != nil {
				r.parent.DeleteServer(addr)
			}
		}
	}()
	r.svr.Addr = addr
	if r.isPprof {
		r.Get("/debug/pprof", func(ctx *Context) {
			pprof.Index(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/pprof/symbol", func(ctx *Context) {
			pprof.Symbol(ctx.w, ctx.GetRequest())
		})

		r.Get("/debug/allocs", func(ctx *Context) {
			pprof.Handler("allocs").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/block", func(ctx *Context) {
			pprof.Handler("block").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/cmdline", func(ctx *Context) {
			pprof.Cmdline(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/goroutine", func(ctx *Context) {
			pprof.Handler("goroutine").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/heap", func(ctx *Context) {
			pprof.Handler("heap").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/mutex", func(ctx *Context) {
			pprof.Handler("mutex").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/profile", func(ctx *Context) {
			pprof.Profile(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/threadcreate", func(ctx *Context) {
			pprof.Handler("threadcreate").ServeHTTP(ctx.w, ctx.GetRequest())
		})
		r.Get("/debug/trace", func(ctx *Context) {
			pprof.Trace(ctx.w, ctx.GetRequest())
		})
	}

	log.Info("start http server on " + addr)
	return r.svr.ListenAndServe()
}

// Stop 停止应用
func (r *Route) Stop() {
	r.cancel()
	r.wg.Wait()
}

// Post 注册 POST 请求路由
func (r *Route) Post(path string, h ...Handler) {
	r.Register(path, MethodPost, h...)
}

// Get 注册 GET 请求路由
func (r *Route) Get(path string, h ...Handler) {
	r.Register(path, MethodGet, h...)
}

// Register 注册路由
func (r *Route) Register(path string, method HTTPMethod, h ...Handler) {
	r.root.Add(path, method, h)
}

// Match 查找路由匹配的处理器
func (r *Route) Match(path, method string) Handles {
	return r.root.Find(path, method)
}

// Group 路由分组
func (r *Route) Group(path string, h ...Handler) *Group {
	return &Group{
		basePath: path,
		middles:  h,
		r:        r,
	}
}

// Use 田间中间件
func (r *Route) Use(h ...Handler) {
	r.middle = append(r.middle, h...)
}

func (r *Route) path(method, path string, h Handler, third ...Handler) {
	method = strings.ToUpper(method)
	if r.paths[method] == nil {
		r.paths[method] = map[string]Handles{}
	}
	handles := make(Handles, 0, len(third)+1)
	if len(third) > 0 {
		handles = append(handles, third...)
	}
	handles = append(handles, h)
	r.paths[method][path] = handles

}

// Error 设置 error 信息
func (r *Route) Error(err error) {
	r.err = err
}

func (r *Route) find(method, path string) (h Handles) {
	method = strings.ToUpper(method)
	if r.paths[method] == nil {
		return nil
	}
	return r.paths[method][path]
}

// New 实例化一个 Router 对象
func New() (r *Route) {
	rou := &Route{
		paths:    map[string]map[string]Handles{},
		ctx:      context.Background(),
		children: make(map[string]*Route),
		isPprof:  true,
		root:     newTree(),
	}
	svr := http.Server{}
	svr.Handler = rou

	rou.svr = &svr

	return rou
}
