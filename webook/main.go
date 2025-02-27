package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func main() {
	//db := initDB()
	//rdb := initRedis()
	//server := initWebServer()
	//
	//u := initUser(db, rdb)
	//u.RegisterRoutes(server)

	initViperV1()

	initPrometheus()
	app := InitWebServer()
	// Consumer 在我设计下，类似于 Web，或者 GRPC 之类的，是一个顶级入口
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	app.cron.Start()

	server := app.web
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，你来了")
	})

	server.Run(":8080")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)
	ctx = app.cron.Stop()
	// 想办法 close ？？
	// 这边可以考虑超时强制退出，防止有些任务，执行特别长的时间
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
	}
}

func initViperReader() {
	viper.SetConfigType("yaml")
	cfg := `
db.mysql:
  dsn: "root:root@tcp(localhost:13316)/webook"

redis:
  addr: "localhost:6379"
`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}
func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		// 通过 webook 和其他使用 etcd 的区别出来
		"http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigFile("yaml")
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name, e.Op)
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func initViperV1() {
	cfile := pflag.String("config", "config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	// 但是它只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 比较好的设计，它会在 e 里面告诉你变更前的数据，和变更后的数据
		// 更好的设计是，它会直接告诉你差异。
		fmt.Println(e.Name, e.Op)
		// 或者自己手动去拿，最后肉眼判断变了没
		fmt.Println(viper.GetString("db.dsn"))
	})
	//viper.SetDefault("db.mysql.dsn",
	//	"root:root@tcp(localhost:3306)/mysql")
	//viper.SetConfigFile("config/dev.yaml")
	//viper.KeyDelimiter("-")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetDefault("db.mysql.dsn", "root:root@tcp(localhost:13316)/webook")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	// 读取配置到 viper 里面，或者你可以理解为加载到内存里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// 一般建议一个 viper 就够了
	//otherViper := viper.New()
	//otherViper.SetConfigName("myjson")
	//otherViper.AddConfigPath("./config")
	//otherViper.SetConfigType("json")
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
