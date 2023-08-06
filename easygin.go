package easygin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/gin-gonic/gin"
)

/*
	use easygin to register controller:
	NOTE: the first parameter must be the type of *gin.Context
	1.bind values from body to struct, you can use the following forms:
		type User struct {
			Id int `json:"id"`
			Username string `json:"username"`
		}
		1.1 func mycontroller(ctx *gin.Context, u User) *Result
		1.2 func mycontroller(ctx *gin.Context, u *User) *Result
	NOTE: using this method will reduce performance by 20%, 2500ns/op --> 3000ns/op

	2.bind values from url(etc.: id=1&username=aabb), you can use the following forms:
		type User struct {
			Id int `form:"id"`
			Username string `form:"username"`
		}
		2.1 func mycontroller(ctx *gin.Context, u User) *Result
		2.2 func mycontroller(ctx *gin.Context, u *User) *Result
	NOTE: using this method will reduce performance by 20%, 2200ns/op --> 2700ns/op

	3. get values from url(etc.: id=1&username=aabb)(supported types are int(int, int8 ...), uint(uint, uint8...), string),
       you can use the following forms:
	   Note: The order of parameters in the function must be consistent with the key value pairs in the url
		3.1 func mycontroller(ctx *gin.Context, id int, username string) *Result
	NOTE: using this method can significantly reduce data acquisition performance, 50ns/op --> 1800ns/op
		  but for the business, this loss can be negligible
*/

type EasyGin struct {
	*gin.Engine
	Server             *http.Server
	signalHandler      func()
	afterCloseHandlers []func()
	maxGraceDuration   time.Duration
}

type RouterGroup struct {
	*gin.RouterGroup
}

func New() *EasyGin {
	return &EasyGin{
		Engine:           gin.New(),
		maxGraceDuration: time.Second * 10,
	}
}

func NewWithEngine(r *gin.Engine) *EasyGin {
	return &EasyGin{Engine: r}
}

func (e *EasyGin) SetMaxGraceDuration(max time.Duration) {
	e.maxGraceDuration = max
}

var elog = log.New(os.Stderr, "EasyGin", log.LstdFlags)

func SetLogOutput(out io.Writer) {
	elog.SetOutput(out)
}

// Handler must be in one of the following forms
// func(ctx *gin.Context) *Result
// func(ctx *gin.Context, u UserType) *Result
// must have one or two parameter
// first param: must be *gin.Context
// second param: must be a struct or a pointer of struct
// return value must be *Result
type Handler interface{}

func (e *EasyGin) GET(relativePath string, handlers ...Handler) {
	e.Engine.GET(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) POST(relativePath string, handlers ...Handler) {
	e.Engine.POST(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) DELETE(relativePath string, handlers ...Handler) {
	e.Engine.DELETE(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) HEAD(relativePath string, handlers ...Handler) {
	e.Engine.HEAD(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) PATCH(relativePath string, handlers ...Handler) {
	e.Engine.PATCH(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) PUT(relativePath string, handlers ...Handler) {
	e.Engine.PUT(relativePath, ginHandlers(handlers...)...)
}

func (e *EasyGin) Group(relativePath string, handlers ...Handler) *RouterGroup {
	group := e.Engine.Group(relativePath, ginHandlers(handlers...)...)
	return &RouterGroup{group}
}

// SetSignalHandler set signal processing functions
// if the signal processing function is set when calling this method,
// the registered afterCloseHandlers will not work
func (e *EasyGin) SetSignalHandler(f func()) {
	e.signalHandler = f
}

// SetAfterCloseHandlers register handlers which will be called after server closed
func (e *EasyGin) SetAfterCloseHandlers(handlers ...func()) {
	e.afterCloseHandlers = append(e.afterCloseHandlers, handlers...)
}

func (e *EasyGin) setupSignal() {
	if e.signalHandler == nil {
		e.signalHandler = func() {
			SetupSignal(func() {
				for _, handler := range e.afterCloseHandlers {
					handler()
				}

				ctx, cancelFunc := context.WithTimeout(context.Background(), e.maxGraceDuration)
				defer cancelFunc()
				if err := e.Server.Shutdown(ctx); err != nil {
					elog.Printf("An error occurs when Server shut:%v", err)
				}
			})
		}
	}

	go e.signalHandler()
}

func (e *EasyGin) ListenAndServe(addr string) error {
	if e.Server == nil {
		e.Server = &http.Server{
			Addr:    addr,
			Handler: e.Engine,
		}
	}

	e.setupSignal()
	return e.Server.ListenAndServe()
}

func (r *RouterGroup) GET(relativePath string, handlers ...Handler) {
	r.RouterGroup.GET(relativePath, ginHandlers(handlers...)...)
}

func (r *RouterGroup) POST(relativePath string, handlers ...Handler) {
	r.RouterGroup.POST(relativePath, ginHandlers(handlers...)...)
}

func (r *RouterGroup) DELETE(relativePath string, handlers ...Handler) {
	r.RouterGroup.DELETE(relativePath, ginHandlers(handlers...)...)
}

func (r *RouterGroup) HEAD(relativePath string, handlers ...Handler) {
	r.RouterGroup.HEAD(relativePath, ginHandlers(handlers...)...)
}

func (r *RouterGroup) PATCH(relativePath string, handlers ...Handler) {
	r.RouterGroup.PATCH(relativePath, ginHandlers(handlers...)...)
}

func (r *RouterGroup) PUT(relativePath string, handlers ...Handler) {
	r.RouterGroup.PUT(relativePath, ginHandlers(handlers...)...)
}

var (
	ginCtxType = reflect.TypeOf(&gin.Context{})
	outType    = reflect.TypeOf(&Result{})
)

const (
	ContentTypeJson = "application/json"
)

func ginHandlers(handlers ...Handler) []gin.HandlerFunc {
	if len(handlers) == 0 {
		return nil
	}
	return pie.Map(handlers, func(handler Handler) gin.HandlerFunc {
		fv := reflect.ValueOf(handler)
		ft := fv.Type()
		if ft.Kind() != reflect.Func {
			panic("handler must be func type")
		}

		// 检查返回值类型，返回值必须只有一个，并且是*Result类型
		if ft.NumOut() == 0 || ft.NumOut() > 1 {
			panic("return value must have one and be *Result type")
		}
		outt := ft.Out(0)
		if outt != outType {
			panic("return value must be *Result type")
		}

		return func(ctx *gin.Context) {
			// 入参可以有0个或多个
			inValues := make([]reflect.Value, 0, ft.NumIn())
			if ft.NumIn() > 0 {
				queryVals := &queryValues{}
				for i := 0; i < ft.NumIn(); i++ {
					in := ft.In(i)
					// 如果当前类型为*gin.Context，则将ctx注入
					if in == ginCtxType {
						inValues = append(inValues, reflect.ValueOf(ctx))
						continue
					}
					// 否则，从gin中取出注入
					val, err := bindParam(in, ctx, queryVals)
					if err != nil {
						return
					}
					inValues = append(inValues, val)
				}
			}

			outVals := fv.Call(inValues)
			result := outVals[0].Interface().(*Result)
			if result == nil {
				return
			}
			ctx.JSON(result.Status, &result.R)

			pool.Put(result)
		}
	})
}

func bindParam(in reflect.Type, ctx *gin.Context, queryVals *queryValues) (reflect.Value, error) {
	isPointer := false
	if in.Kind() == reflect.Pointer {
		in = in.Elem()
		isPointer = true
	}
	inVal := reflect.New(in)

	switch in.Kind() {
	case reflect.Struct:
		// bind from url
		if ctx.Request.ContentLength == 0 {
			err := ctx.BindQuery(inVal.Interface())
			if err != nil {
				return reflect.Value{}, err
			}

		} else if ctx.ContentType() == ContentTypeJson { // bind json from body
			err := ctx.BindJSON(inVal.Interface())
			if err != nil {
				return reflect.Value{}, err
			}

		} else {
			err := ctx.Bind(inVal.Interface())
			if err != nil {
				return reflect.Value{}, err
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.String:
		if !queryVals.inited {
			if ctx.Request.URL.RawQuery == "" {
				return reflect.Value{}, errors.New("query is empty")
			}
			err := parseQuery(ctx.Request.URL.RawQuery, queryVals)
			if err != nil {
				return reflect.Value{}, err
			}
		}
		if queryVals.index >= len(queryVals.keys) {
			return reflect.Value{}, errors.New("query is empty")
		}
		key := queryVals.keys[queryVals.index]
		queryVals.index++
		if key == "" {
			return reflect.Value{}, errors.New("get query key error")
		}

		v := queryVals.kvs[key]
		switch in.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			inVal.Elem().SetInt(int64(n))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			inVal.Elem().SetUint(uint64(n))
		case reflect.String:
			inVal.Elem().SetString(v[0])
		}
	}

	if isPointer {
		return inVal, nil
	}

	return inVal.Elem(), nil
}

type queryValues struct {
	kvs    map[string][]string
	keys   []string
	index  int
	inited bool
}

func parseQuery(query string, values *queryValues) (err error) {
	kvsc := strings.Count(query, "&") + 1
	values.kvs = make(map[string][]string, kvsc)
	values.keys = make([]string, 0, kvsc)

	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}

		vals, ok := values.kvs[key]
		if !ok {
			values.keys = append(values.keys, key)
		}
		values.kvs[key] = append(vals, value)
	}

	values.inited = true

	return
}
