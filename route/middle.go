package route

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"github.com/HiData-xyz/hit/log"

	"github.com/dgrijalva/jwt-go"
)

// CORS 运行跨域
func CORS(ctx *Context) {
	log.Info(fmt.Sprintf("request header: %+v", ctx.r.Header.Get("origin")))
	ctx.w.Header().Set("Access-Control-Allow-Origin", ctx.r.Header.Get("Origin"))
	if ctx.r.Method == http.MethodOptions {
		ctx.w.Header().Set("Access-Control-Allow-Headers", strings.Join(ctx.r.Header["Access-Control-Request-Headers"], ","))
		ctx.w.Header().Set("Access-Control-Allow-Methods", strings.Join(ctx.r.Header["Access-Control-Request-Method"], ","))
		ctx.w.Header().Set("Access-Control-Max-Age", "1728000")
		ctx.w.WriteHeader(204)
		ctx.Stop()
	}
	return
}

var ErrInvalidToken = errors.New("token checked is failed")

// Token 中间件
func Token(header string, base string, v interface{}) Handler {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	v = reflect.New(typ).Interface()

	return func(ctx *Context) {
		token := ctx.r.Header.Get(header)
		if token == "" {
			ctx.EJSON(http.StatusUnauthorized, ErrInvalidToken.Error())
			ctx.Stop()
			return
		}

		_v := v.(jwt.Claims)
		_token, err := jwt.ParseWithClaims(token, _v, func(token *jwt.Token) (v interface{}, err error) {
			return []byte(base), nil
		})
		if err != nil {
			ctx.EJSON(http.StatusBadRequest, err.Error())
			ctx.Stop()
			return
		}
		if !_token.Valid {
			ctx.EJSON(http.StatusUnauthorized, ErrInvalidToken.Error())
			ctx.Stop()
			return
		}
		ctx.SetValue("token", _token.Claims)
	}
}
