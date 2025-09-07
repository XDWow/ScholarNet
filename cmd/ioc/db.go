package ioc

import (
	"github.com/XD/ScholarNet/cmd/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// 通过全局变量来配置
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))

	// 在初始化的地方，定义一个内部结构体，用来接收全部相关的配置。
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	// 这样就定义好了默认值
	var cfg = Config{
		DSN: "root:root@tcp(localhost:13316)/webook_default",
	}
	// remote 不支持 key 的切割:db.mysql
	err := viper.UnmarshalKey("db", &cfg)
	//dsn := viper.GetString("db.mysql")
	//println(dsn)
	//if err != nil {
	//	panic(err)
	//}
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
