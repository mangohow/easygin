package easygin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"testing"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func TestEasyGin(t *testing.T) {
	easyGin := New()
	easyGin.GET("/", func(ctx *gin.Context, user User) *Result {
		fmt.Println(user)
		return nil
	})

	easyGin.GET("/pointer", func(ctx *gin.Context, user *User) *Result {
		fmt.Println(user)
		return nil
	})

	easyGin.POST("/post", func(ctx *gin.Context, user User) *Result {
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

func TestFunction(t *testing.T) {
	easyGin := New()
	easyGin.POST("/post", func(ctx *gin.Context, user *User) *Result {
		return Ok(user, "ok")
	})

	group := easyGin.Group("/api")
	group.POST("/post", func(ctx *gin.Context, user *User) *Result {
		return Ok(user, "ok")
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
	ctx.Request.Body = Data(data)
	return ctx
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
	f := ginHandlers(func(ctx *gin.Context, user *User) *Result {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}

func BenchmarkReflect(b *testing.B) {
	ctx := ginContext()
	f := ginHandlers(func(ctx *gin.Context, user User) *Result {
		return nil
	})[0]

	for i := 0; i < b.N; i++ {
		f(ctx)
	}
}
