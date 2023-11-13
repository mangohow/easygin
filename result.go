package easygin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// Response controller返回值, 不要自己创建, 一定要调用函数来获取
type Response struct {
	R      RespValue
	Status int
}

type RespValue struct {
	RespError
	Data interface{} `json:"data"`
}

const (
	jsonData    = `{"data":`
	jsonCode    = `,"code":`
	jsonMessage = `,"message":"`
	jsonEnd     = `"}`
	jsonLen     = len(`{"data":, "code":,"message":""}`)
)

func (r *RespValue) MarshalJSON() ([]byte, error) {
	// {"data":null,"code":2,"message":"failed"}
	bs, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}
	num := strconv.Itoa(r.Code())

	buffer := bytes.NewBuffer(nil)
	buffer.Grow(jsonLen + len(bs) + len(num) + len(r.Message()))
	buffer.WriteString(jsonData)
	buffer.Write(bs)
	buffer.WriteString(jsonCode)
	buffer.WriteString(num)
	buffer.WriteString(jsonMessage)
	buffer.WriteString(r.Message())
	buffer.WriteString(jsonEnd)

	return buffer.Bytes(), nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Response{}
	},
}

func NewResponse(status int, data interface{}, respErr RespError) *Response {
	res := pool.Get().(*Response)
	res.Status = status
	res.R.RespError = respErr
	res.R.Data = data

	return res
}

func Ok() *Response {
	return NewResponse(http.StatusOK, nil, RespSuccess)
}

func OkData(data interface{}) *Response {
	return NewResponse(http.StatusOK, data, RespSuccess)
}

func OkCode(code int) *Response {
	return NewResponse(http.StatusOK, nil, NewError(code, "success"))
}

func OkCodeData(code int, data interface{}) *Response {
	return NewResponse(http.StatusOK, data, NewError(code, "success"))
}

func Fail(respErr RespError) *Response {
	return FailData(respErr, nil)
}

func FailData(respErr RespError, data interface{}) *Response {
	return NewResponse(http.StatusOK, data, respErr)
}

func Error(status int) *Response {
	return NewResponse(status, nil, nil)
}
