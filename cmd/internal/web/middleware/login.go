package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	// 注册一下这个类型
	gob.Register(time.Now())
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(c)
		if sess.Get("userId") == nil {
			//未登录，中断，不要往后执行，也就是不要执行后面的业务逻辑
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		updata_time := sess.Get("updata_time")
		now := time.Now().UnixMilli()
		// 说明刚登录，还没刷新过，需要初始化updata_time
		if updata_time == nil {
			sess.Set("updata_time", now)
			sess.Save()
			return
		}

		//updata_time存在，需要判断过了多久
		updata_timeval, _ := updata_time.(int64)
		if now-updata_timeval > 60*1000 {
			sess.Set("updata_time", now)
			sess.Save()
			sess.Options(sessions.Options{
				MaxAge: 60,
			})
		}
	}
}
