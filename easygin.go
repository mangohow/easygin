package easygin

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"reflect"
	"time"
)

/*
	Using reflection to modify the handler has a performance
	loss of approximately 20%, but compared to business processing, this loss is small
*/

type EasyGin struct {
	*gin.Engine
	Server            *http.Server
	signalHandler     func()
	afterCloseHandler []func()
}

type RouterGroup struct {
	*gin.RouterGroup
}

func New() *EasyGin {
	return &EasyGin{
		Engine: gin.New(),
	}
}

func NewWithEngine(r *gin.Engine) *EasyGin {
	return &EasyGin{Engine: r}
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
	e.afterCloseHandler = append(e.afterCloseHandler, handlers...)
}

func (e *EasyGin) setupSignal() {
	if e.signalHandler == nil {
		e.signalHandler = func() {
			SetupSignal(func() {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
				defer cancelFunc()
				if err := e.Server.Shutdown(ctx); err != nil {
					log.Printf("An error occurs when Server shut:%v", err)
				}

				for _, handler := range e.afterCloseHandler {
					handler()
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

		// check the number of parameters
		if ft.NumIn() <= 0 || ft.NumIn() > 2 {
			panic("handler must have one or two parameter")
		}

		// check whether the first parameter is of type *gin.Context
		in0 := ft.In(0)
		if in0 != ginCtxType {
			panic("first parameter must be *gin.Context type")
		}

		// check the return value type, the type must be *Result
		// and return value count must be one
		if ft.NumOut() == 0 || ft.NumOut() > 1 {
			panic("return value must have one and be *Result type")
		}

		outt := ft.Out(0)
		if outt != outType {
			panic("return value must be *Result type")
		}

		return func(ctx *gin.Context) {
			inValues := make([]reflect.Value, 0, ft.NumIn())
			inValues = append(inValues, reflect.ValueOf(ctx))

			for i := 1; i < ft.NumIn(); i++ {
				in := ft.In(i)
				val, err := bindParam(in, ctx)
				if err != nil {
					fmt.Println(err)
					return
				}
				inValues = append(inValues, val)
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

func bindParam(in reflect.Type, ctx *gin.Context) (reflect.Value, error) {
	isPointer := false
	if in.Kind() == reflect.Pointer {
		in = in.Elem()
		isPointer = true
	}
	inVal := reflect.New(in)

	switch in.Kind() {
	case reflect.Struct:
		err := ctx.Bind(inVal.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		if ctx.Request.Method == http.MethodGet && ctx.ContentType() == ContentTypeJson {
			err := ctx.BindJSON(inVal.Interface())
			if err != nil {
				fmt.Println(err)
				return reflect.Value{}, err
			}
		}

	}

	if isPointer {
		return inVal, nil
	}

	return inVal.Elem(), nil
}
