package domain

type AsyncSms struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
	// 最大重试次数
	RetryMax int
	// 服务商重试策略
	Strategy string
	// 去重用的
	BizType string
	BizID   string
}
