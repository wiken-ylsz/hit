package route

// Group 实现路由分组
type Group struct {
	basePath string
	middles  Handles
	r        *Route
}

// Post 注册POST请求
func (g *Group) Post(path string, h ...Handler) {
	path = g.basePath + path
	h = append(g.middles, h...)
	g.r.Post(path, h...)
}

// Get 注册Get请求
func (g *Group) Get(path string, h ...Handler) {
	path = g.basePath + path
	h = append(g.middles, h...)
	g.r.Get(path, h...)
}

// New 基于当前组, 创建新的路由分组
func (g *Group) New(path string, h ...Handler) *Group {
	return &Group{
		basePath: g.basePath + path,
		middles:  append(g.middles, h...),
		r:        g.r,
	}
}
