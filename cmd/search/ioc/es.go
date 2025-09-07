package ioc

import (
	"fmt"
	"github.com/XD/ScholarNet/cmd/search/repository/dao"
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"time"
)

// InitESClient 读取配置文件，进行初始化ES客户端
func InitESClient() *elastic.Client {
	type Config struct {
		Url   string `json:"url"`
		Sniff bool   `json:"sniff"`
	}
	var cfg Config
	err := viper.UnmarshalKey("es", &cfg)
	if err != nil {
		panic(fmt.Errorf("读取 ES 配置文件失败: %w", err))
	}
	const timeout = time.Second * 100
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Url),
		elastic.SetSniff(cfg.Sniff),
		elastic.SetHealthcheckTimeout(timeout),
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	err = dao.InitES(client)
	if err != nil {
		panic(err)
	}
	return client
}
