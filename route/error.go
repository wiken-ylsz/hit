package route

import "errors"

// 常用错误
var (
	ErrReadRequestBodyFail = errors.New("read from request body failed")
)
