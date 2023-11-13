package easygin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

/*
goos: windows
goarch: amd64
pkg: github.com/mangohow/easygin
cpu: Intel(R) Core(TM) i7-7700 CPU @ 3.60GHz
BenchmarkNormal-8                         492842              2492 ns/op
BenchmarkReflectPointer-8                 418363              2878 ns/op
BenchmarkReflect-8                        428484              2983 ns/op
BenchmarkNormalQuery-8                  23289032                52.07 ns/op
BenchmarkReflectQuery-8                   668437              1825 ns/op
BenchmarkStructQuery-8                    559951              2256 ns/op
BenchmarkStructReflectQuery-8             439838              2730 ns/op
PASS
ok      github.com/mangohow/easygin     9.822s
*/

type User struct {
	Id       int    `json:"id" form:"id"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Email    string `json:"email" form:"email"`
}

func TestEasyGin(t *testing.T) {
	easyGin := New()
	easyGin.GET("/hello", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.GET("/noparam", func() *Response {
		fmt.Println("no param")
		return nil
	})

	easyGin.GET("/noctx", func(ctx *gin.Context) *Response {
		fmt.Println(ctx)
		return nil
	})

	easyGin.GET("/helloworld", func(user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.GET("/hello/world", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.GET("/query", func(ctx *gin.Context, id int, username string) *Response {
		fmt.Println(id, username)
		return nil
	})

	easyGin.GET("/queryPointer", func(ctx *gin.Context, id *int, username *string) *Response {
		fmt.Println(*id, *username)
		return nil
	})

	easyGin.POST("/", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.PUT("/", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.DELETE("/", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.HEAD("/", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.GET("/pointer", func(ctx *gin.Context, user *User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.POST("/post", func(ctx *gin.Context, user User) *Response {
		fmt.Println(user)
		return nil
	})

	easyGin.Engine.GET("/test", func(ctx *gin.Context) {
		user := &User{}
		err := ctx.BindJSON(user)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(user)
	})

	easyGin.Run(":8080")

}

func TestMarshalJson(t *testing.T) {
	resp := FailData(NewError(2, "fail"), "failed")
	bytes, err := json.Marshal(&resp.R)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bytes))

	resp = FailData(NewError(2, "fail"), nil)
	bytes, err = json.Marshal(&resp.R)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bytes))

	type UserInfo struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
	}

	resp = FailData(NewError(2, "fail"), &UserInfo{Id: 1, Username: "aabb"})
	bytes, err = json.Marshal(&resp.R)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bytes))
}

var UsernameInvalidErr = &RespErrorImpl{
	Codee:    1,
	Messagee: "username invalid",
}

func checkUsername() error {
	return UsernameInvalidErr
}

func TestFunction(t *testing.T) {
	easyGin := New()

	easyGin.GET("/test", func(ctx *gin.Context) *Response {
		err := checkUsername()
		if IsRespError(err) {
			return Fail(AsRespError(err))
		}
		if err != nil {
			return Fail(NewFromError(err))
		}

		return OkData("ok")
	})

	easyGin.POST("/post", func(ctx *gin.Context, user *User) *Response {
		return OkData(user)
	})

	group := easyGin.Group("/api")
	group.POST("/post", func(ctx *gin.Context, user *User) *Response {
		return OkData(user)
	})

	easyGin.SetAfterCloseHandlers(func() {
		fmt.Println("just a test")
	})

	err := easyGin.ListenAndServe(":8080")

	if err != nil {
		fmt.Println(err)
	}

}

type Data []byte

func (d Data) Read(p []byte) (n int, err error) {
	copy(p, d)
	return len(d), nil
}

func (d Data) Close() error {
	return nil
}

func TestContext(t *testing.T) {
	ctx := ginContext()
	user := &User{}
	err := ctx.Bind(&user)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user)
}

func ginContext() *gin.Context {
	ctx := &gin.Context{}
	ctx.Request = &http.Request{}
	ctx.Request.Method = http.MethodPost
	ctx.Request.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	u := &User{
		Id:       100,
		Username: "aaaaaaaaaaaaaaaaaaaaaaaa",
		Password: "wwwwwwwwwwwwwwwwwwwwww",
		Email:    "eeeeeeeeeeeeeeeeeeeeee",
	}
	data, _ := json.Marshal(u)
	ctx.Request.Header.Set("Content-Length", strconv.Itoa(len(data)))
	ctx.Request.ContentLength = int64(len(data))
	ctx.Request.Body = Data(data)
	return ctx
}

func ginQueryContext() *gin.Context {
	ctx := &gin.Context{}
	ctx.Request = &http.Request{}
	ctx.Request.Method = http.MethodGet
	ctx.Request.URL = &url.URL{
		RawQuery: "id=1&username=aabb&password=ccdd&username=1234&email=aa@bb.com",
	}

	return ctx
}

func TestParseQuery(t *testing.T) {
	query := "id=1&username=aabb&password=ccdd&email=aa@bb.com"
	queryVals := &queryValues{}
	err := parseQuery(query, queryVals)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(queryVals)
}

func BenchmarkNormal(b *testing.B) {
	ctx := ginContext()
	f := func(ctx *gin.Context) {
		user := User{}
		_ = ctx.Bind(&user)
	}
	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkReflectPointer(b *testing.B) {
	ctx := ginContext()
	f := ginHandlers(func(ctx *gin.Context, user *User) *Response {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkReflect(b *testing.B) {
	ctx := ginContext()
	f := ginHandlers(func(ctx *gin.Context, user User) *Response {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkNormalQuery(b *testing.B) {
	ctx := ginQueryContext()
	f := func(ctx *gin.Context) {
		ctx.Query("id")
		ctx.Query("username")
		ctx.Query("password")
		ctx.Query("email")
	}

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkReflectQuery(b *testing.B) {
	ctx := ginQueryContext()
	f := ginHandlers(func(ctx *gin.Context, id int, username, password, email string) *Response {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkStructQuery(b *testing.B) {
	ctx := ginQueryContext()
	f := func(ctx *gin.Context) {
		u := &User{}
		ctx.BindQuery(u)
	}

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkStructReflectQuery(b *testing.B) {
	ctx := ginQueryContext()
	f := ginHandlers(func(ctx *gin.Context, user *User) *Response {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

type Resp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestResponseAndError(t *testing.T) {
	server := NewWithEngine(gin.New())
	server.GET("/test", func(ctx *gin.Context) *Response {
		resp, err := service()
		if err != nil {
			t.Log("err:", err)
			if IsRespError(err) {
				return Fail(AsRespError(err))
			}
			return Fail(NewFromError(err))
		}

		return OkData(resp)
	})

	server.ListenAndServe(":8080")
}

func service() (*Resp, error) {
	n := rand.Intn(3)
	switch n {
	case 1:
		return nil, NewError(1, "test")
	case 2:
		return nil, errors.New("internal error")
	}

	return &Resp{
		Id:   1,
		Name: "ape",
	}, nil
}
