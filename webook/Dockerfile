#基础镜像，整个镜像是在基础镜像上叠加
FROM ubuntu:20.04
# 把编译后的打包进来这个镜像，放在工作目录/app，这个可以随便换
COPY webook /app/webook
WORKDIR /app

ENTRYPOINT ["/app/webook"]