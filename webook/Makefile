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