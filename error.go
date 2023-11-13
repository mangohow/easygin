package easygin

import "fmt"

type RespError interface {
	error
	Message() string
	Code() int
}

// RespErrorImpl
// e is added at the end because we need to export the field
// and it cannot have the same name as the function
type RespErrorImpl struct {
	Codee    int    `json:"code"`
	Messagee string `json:"message"`
}

func (e *RespErrorImpl) Error() string {
	return fmt.Sprintf("[%d]%s", e.Codee, e.Messagee)
}

func (e *RespErrorImpl) Message() string {
	return e.Messagee
}

func (e *RespErrorImpl) Code() int {
	return e.Codee
}

func NewError(code int, msg string) RespError {
	return &RespErrorImpl{
		Codee:    code,
		Messagee: msg,
	}
}

func NewErrorf(code int, format string, args ...interface{}) RespError {
	return &RespErrorImpl{
		Codee:    code,
		Messagee: fmt.Sprintf(format, args...),
	}
}

func NewFromError(err error) RespError {
	return &RespErrorImpl{
		Codee:    UnknownErrorCode,
		Messagee: err.Error(),
	}
}

func IsRespError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(RespError)
	return ok
}

func AsRespError(err error) RespError {
	if err == nil {
		return nil
	}
	re, ok := err.(RespError)
	if !ok {
		return nil
	}
	return re
}

func IsSuccess(err RespError) bool {
	return err.Code() == SuccessCode
}

const (
	UnknownErrorCode = -1
	SuccessCode      = 0
)

var (
	RespSuccess = RespError(&RespErrorImpl{
		Codee:    SuccessCode,
		Messagee: "success",
	})
)
