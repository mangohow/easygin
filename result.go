package easygin

import (
	"net/http"
	"sync"
)

// Result controller返回值, 不要自己创建, 一定要调用函数来获取
type Result struct {
	R      Response
	Status int
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	Success = iota
	UnknownError
)

var pool = sync.Pool{
	New: func() interface{} {
		return &Result{}
	},
}

func NewResult(status, code int, data interface{}, message string) *Result {
	res := pool.Get().(*Result)
	res.Status = status
	res.R.Code = code
	res.R.Data = data
	res.R.Message = message

	return res
}

func Ok(data interface{}, message string) *Result {
	return NewResult(http.StatusOK, Success, data, message)
}

func OkNoData(message string) *Result {
	return NewResult(http.StatusOK, Success, nil, message)
}

func Error(status int, code int, message string) *Result {
	return NewResult(status, code, nil, message)
}
