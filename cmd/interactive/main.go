package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

func main() {
	initViperV1()
	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		// http，8081 端口
		err := app.webAdmin.Start()
		log.Println(err)
	}()
	err := app.server.Serve()
	log.Println(err)
}

//func main() {
//	server := grpc2.NewServer()
//	intrSvc := &grpc.InteractiveServiceServer{}
//	intrv1.RegisterInteractiveServiceServer(server, intrSvc)
//	// 监听 8090 端口，你可以随便写
//	l, err := net.Listen("tcp", ":8090")
//	if err != nil {
//		panic(err)
//	}
//	// 这边会阻塞，类似与 gin.Run
//	err = server.Serve(l)
//	log.Println(err)
//}

func initViperV1() {
	cfile := pflag.String("config", "config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 比较好的设计，它会在 in 里面告诉你变更前的数据，和变更后的数据
		// 更好的设计是，它会直接告诉你差异
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
