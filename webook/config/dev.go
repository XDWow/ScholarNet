//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		//本地连接
		DSN: "root:root@tcp(localhost:30001)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:30003",
	},
}
