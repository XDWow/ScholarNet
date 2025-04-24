.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
#	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
#windows的方法
	@set GOOS=linux
	@set GOARCH=arm
	@go build -tags=k8s -o webook .
	@docker rmi -f jj/webook:v0.0.1
	@docker build -t jj/webook:v0.0.1 .

mockgen -source webook/internal/repository/user.go -destination webook/internal/repository/mocks/user.mock.go -package repomocks
mockgen -source webook\internal\service\article.go -destination webook\internal\service\mocks\article.mock.go -package svcmocks
mockgen -source webook\internal\repository\article\article_author.go -destination webook\internal\repository\article\mocks\article_author.mock.go -package repositorymocks
mockgen -source webook\internal\repository\article\article_reader.go -destination webook\internal\repository\article\mocks\article_reader.mock.go -package repositorymocks

protoc --go_out=../../gen/payment/v1 --go_opt=paths=source_relative --go-grpc_out=../../gen/payment/v1 --go-grpc_opt=paths=source_relative payment.proto
