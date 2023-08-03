package easygin

import (
	"net/http"
	"sync"
)

type CodeMessager func(int) string

var cm CodeMessager

// SetCodeMessager 设置一个CodeMessager，该函数可以从code获取message
func SetCodeMessager(messager CodeMessager) {
	cm = messager
}

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
	SuccessCode = iota
	FailCode
	ErrorCode
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

func Ok(data interface{}) *Result {
	return NewResult(http.StatusOK, SuccessCode, data, "success")
}

func Fail(code int) *Result {
	return FailWithData(nil, code)
}

func FailWithData(data interface{}, code int) *Result {
	if cm != nil {
		return NewResult(http.StatusOK, code, data, cm(code))
	}
	return NewResult(http.StatusOK, FailCode, data, "failed")
}

func Error(status int, code int) *Result {
	if cm != nil {
		return NewResult(status, code, nil, cm(code))
	}
	return NewResult(status, ErrorCode, nil, "error")
}
