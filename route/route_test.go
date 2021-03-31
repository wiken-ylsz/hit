package route_test

import (
	"testing"
	"github.com/HiData-xyz/hit/route"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutePath(t *testing.T) {
	Convey("测试路由注册", t, func() {
		Convey("注册POST请求", func() {
			routePost := []struct {
				path string
				h    route.Handler
			}{
				{"/user", func(ctx *route.Context) {}},
				{"/user?id=10", func(ctx *route.Context) {}},
			}
			r := route.New()
			for _, val := range routePost {
				r.Post(val.path, val.h)
			}
			for _, val := range routePost {
				h := r.Match(val.path, route.MethodPost)
				So(h, ShouldHaveLength, 1)

				h = r.Match(val.path, route.MethodGet)
				So(h, ShouldHaveLength, 0)
				So(h, ShouldBeZeroValue)
			}
		})

		Convey("注册GET请求", func() {
			routePost := []struct {
				path string
				h    route.Handler
			}{
				{"/user", func(ctx *route.Context) {}},
			}
			r := route.New()
			for _, val := range routePost {
				r.Get(val.path, val.h)
			}
			for _, val := range routePost {
				h := r.Match(val.path, route.MethodGet)
				So(h, ShouldHaveLength, 1)

				h = r.Match(val.path, route.MethodPost)
				So(h, ShouldHaveLength, 0)
				So(h, ShouldBeZeroValue)
			}
		})

		Convey("RESTFUL路由注册", func() {
			routePost := []struct {
				path string
				h    route.Handler
			}{
				{"/user", func(ctx *route.Context) {}},
			}
			r := route.New()
			for _, val := range routePost {
				r.Get(val.path, val.h)
				r.Post(val.path, val.h)
			}
			for _, val := range routePost {
				h := r.Match(val.path, route.MethodGet)
				So(h, ShouldHaveLength, 1)

				h = r.Match(val.path, route.MethodPost)
				So(h, ShouldHaveLength, 1)
			}
		})

		Convey("路由未注册", func() {
			routePost := []struct {
				path string
				h    route.Handler
			}{
				{"/user", func(ctx *route.Context) {}},
			}
			r := route.New()
			for _, val := range routePost {
				r.Get(val.path, val.h)
				r.Post(val.path, val.h)
			}
			for _, val := range routePost {
				h := r.Match(val.path+val.path, route.MethodGet)
				So(h, ShouldBeZeroValue)
			}
		})

		Convey("请求方法不匹配", func() {
			routePost := []struct {
				path         string
				registMethod string
				matchMenthod string
				h            route.Handler
			}{
				{"/user", route.MethodGet, route.MethodPost, func(ctx *route.Context) {}},
			}
			r := route.New()
			for _, val := range routePost {
				switch val.registMethod {
				case route.MethodGet:
					r.Get(val.path, val.h)
				case route.MethodPost:
					r.Post(val.path, val.h)
				}
			}
			for _, val := range routePost {
				h := r.Match(val.path, val.matchMenthod)
				So(h, ShouldBeZeroValue)
			}
		})
	})
}
