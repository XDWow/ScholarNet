package web

import (
	"bytes"
	"encoding/json"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/service"
	svcmocks "github.com/LXD-c/basic-go/webook/internal/service/mocks"
	"github.com/LXD-c/basic-go/webook/ioc"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Edit(t *testing.T) {
	type fields struct {
		svc service.ArticleService
	}
	type args struct {
		context *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ArticleHandler{
				svc: tt.fields.svc,
			}
			h.Edit(tt.args.context)
		})
	}
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctel *gomock.Controller) service.ArticleService

		reqBody string

		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			l := ioc.InitLogger()
			h := NewArticleHandler(testCase.mock(ctrl), l)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/article/publish", bytes.NewBuffer([]byte(testCase.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, testCase.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var res Result
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, testCase.wantRes, res)
		})
	}
}
