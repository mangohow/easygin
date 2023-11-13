# EasyGin

EasyGin is a Go library that provides a wrapper for the Gin framework, aiming to make working with the Gin framework easier and more convenient. By offering additional features and functionalities, EasyGin helps you build web applications based on Gin more quickly.

## Features

- **Automatic Parameter Parsing and Injection**: EasyGin automatically parses incoming requests and injects the parameters into your controller functions, allowing you to focus on writing business logic without worrying about parsing request data.
- **Unified Error Handling**: EasyGin offers a unified error handling mechanism, allowing you to handle and respond to errors in a consistent and structured way across your application.
- **Graceful Server Shutdown**: EasyGin provides a graceful server shutdown mechanism, ensuring that active connections are completed before the server shuts down, preventing data loss or abrupt termination.
- **Runtime Profile Collection**: EasyGin includes runtime profiling functionality, allowing you to collect performance profiles of your application during runtime for analysis and optimization.

## Examples

```go
server := easygin.NewWithEngine(gin.New())
server.GET("/test", func(ctx *gin.Context) *easygin.Response {
    resp, err := service()
    if err != nil {
        if easygin.IsRespError(err) {
            return easygin.Fail(easygin.AsRespError(err))
        }
        return easygin.Fail(easygin.NewFromError(err))
    }

    return easygin.OkData(resp)
})

server.ListenAndServe(":8080")
```

&nbsp;

request parameters can also be added to the controller, and easygin will automatically parse and inject the data in the request
```go
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

server.GET("/test", func(ctx *gin.Context, user *User) *Response {
    ...
})
```


## Installation

Use `go mod` for dependency management:

```shell
go get github.com/mangohow/easygin