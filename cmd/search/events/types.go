package events

// 暴露给 main 用的接口
type Consumer interface {
	Start() error
}
