package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	//db := initDB()
	//rdb := initRedis()
	//server := initWebServer()
	//
	//u := initUser(db, rdb)
	//u.RegisterRoutes(server)
	server := InitWebServer()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "你好！")
	})

	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	//login := middleware.LoginMiddlewareBuilder{}
	//Gin Session 存储的实现

	//1.基于cookie
	//sessions.Sessions()方法用于生成一个会话中间件，看源码，cookie会存储ssid，当用户发起请求时，浏览器会自动将存储的 Cookie（包括会话标识）发送给服务器。
	//会话中间件通过解析 Cookie 中的会话标识，来检索服务器端存储的会话数据（存在store中)。
	//store := cookie.NewStore([]byte("secret"))

	//2.单机单实例部署，考虑memstore，内存实现
	//store := memstore.NewStore([]byte("ixJX摩4N7G!zM9U5LkW&3$vVnP"),
	//	[]byte("7rW3!z1G0fX5Cp9Z,oK@2mB8"),
	//)

	//store, err := redis.NewStore(16, "tcp",
	//	config.Config.Redis.Addr, "",
	//	[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgK"),
	//	[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgA"))
	//if err != nil {
	//	panic(err)
	//}

	//server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
	//return server
	return server
}
