//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		//本地连接
		DSN: "root:root@tcp(webook-mysql:11309)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:6379",
	},
}
