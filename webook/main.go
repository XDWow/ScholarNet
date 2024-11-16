package main

import (
	"gitee.com/geekbang/basic-go/webook/config"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "你好！")
	})
	server.Run(":8080")
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	//使用use方法注册middleware，第一个是用于解决 CORS 的 middleware
	server.Use(cors.New(cors.Config{
		//是否允许带上用户认证信息（比如 cookie）
		AllowCredentials: true,
		AllowHeaders:     []string{"content-type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		//哪些来源是允许的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "live.webook.com")
		},
		MaxAge: 12 * time.Hour,
	}), func(c *gin.Context) {
		println("这是我的 Middleware")
	})

	//login := middleware.LoginMiddlewareBuilder{}
	login := middleware.LoginJwtMiddlewareBuilder{}
	//Gin Session 存储的实现

	//1.基于cookie
	//sessions.Sessions()方法用于生成一个会话中间件，看源码，cookie会存储ssid，当用户发起请求时，浏览器会自动将存储的 Cookie（包括会话标识）发送给服务器。
	//会话中间件通过解析 Cookie 中的会话标识，来检索服务器端存储的会话数据（存在store中)。
	//store := cookie.NewStore([]byte("secret"))

	//2.单机单实例部署，考虑memstore，内存实现
	//store := memstore.NewStore([]byte("ixJX摩4N7G!zM9U5LkW&3$vVnP"),
	//	[]byte("7rW3!z1G0fX5Cp9Z,oK@2mB8"),
	//)

	store, err := redis.NewStore(16, "tcp",
		config.Config.Redis.Addr, "",
		[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgK"),
		[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgA"))
	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
	return server
}
