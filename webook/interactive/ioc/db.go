package ioc

import (
	"github.com/LXD-c/basic-go/webook/interactive/repository/dao"
	"github.com/LXD-c/basic-go/webook/pkg/gormx/connpool"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSRC() SrcDB { return InitDB("src") }
func InitDST() DstDB { return InitDB("dst") }

// 这个是业务用的，支持双写的 DB
func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: pool,
	}))
	if err != nil {
		panic(err)
	}
	return db
}

func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	return connpool.NewDoubleWritePool(src.ConnPool, dst.ConnPool, pattern)
}

func InitDB(key string) *gorm.DB {
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
	err := viper.UnmarshalKey("db."+key, &cfg)
	//dsn := viper.GetString("db.mysql")
	//println(dsn)
	//if err != nil {
	//	panic(err)
	//}
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
