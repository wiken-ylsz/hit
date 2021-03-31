package route_test

import (
	"testing"
	"github.com/HiData-xyz/hit/route"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouteGroup(t *testing.T) {
	Convey("测试分组路由", t, func() {
		groupPaths := []struct {
			basePath string
			h        route.Handler
		}{
			{"/user", func(ctx *route.Context) {}},
		}
		paths := []*struct {
			fulpath    string
			method     string
			path       string
			lenHandler int
			h          route.Handler
		}{
			{"", route.MethodPost, "/login", 0, func(ctx *route.Context) {}},
			{"", route.MethodPost, "/user/login", 0, func(ctx *route.Context) {}},
		}

		r := route.New()
		for _, val := range groupPaths {
			g := r.Group(val.basePath, val.h)
			for _, v := range paths {
				v.fulpath = val.basePath + v.path
				v.lenHandler = 2
				switch v.method {
				case route.MethodPost:
					g.Post(v.path, v.h)
				case route.MethodGet:
					g.Get(v.path, v.h)
				}
			}
		}

		for _, v := range paths {
			h := r.Match(v.fulpath, v.method)
			So(h, ShouldHaveLength, v.lenHandler)
		}
	})
}
