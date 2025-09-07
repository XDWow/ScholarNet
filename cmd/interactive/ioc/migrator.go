package ioc

import (
	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/interactive/repository/dao"
	"github.com/XD/ScholarNet/cmd/pkg/ginx"
	"github.com/XD/ScholarNet/cmd/pkg/gormx/connpool"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/migrator/events"
	"github.com/XD/ScholarNet/cmd/pkg/migrator/events/fixer"
	"github.com/XD/ScholarNet/cmd/pkg/migrator/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

// 要初始化数据迁移相关的
// 从上往下，顺藤摸瓜，顶级入口有：
// wedAdmin
// 生产者，消费者
const topic = "migrator_interactive"

func InitMigratorWeb(
	src SrcDB,
	dst DstDB,
	l logger.LoggerV1,
	pool *connpool.DoubleWritePool,
	producer events.Producer) *ginx.Server {
	// 在这里，有多少张表，就初始化多少个 scheduler
	intrSch := scheduler.NewScheduler[dao.Interactive](src, dst, l, pool, producer)
	engine := gin.Default()
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "LXD",
		Subsystem: "webook_intr_admin",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	intrSch.RegisterRoutes(engine.Group("/migrator"))
	//intrSch.RegisterRoutes(engine.Group("/migrator/interactive"))
	addr := viper.GetString("migrator.web.addr")
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}

func InitMigratorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitFixDataConsumer(l logger.LoggerV1,
	src SrcDB,
	dst DstDB,
	client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, src, dst, topic)
	if err != nil {
		panic(err)
	}
	return res
}
