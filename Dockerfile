FROM docker.m.daocloud.io/library/golang:1.26.3-bookworm

WORKDIR /app
ENV GOTOOLCHAIN=local \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build ./...
ENTRYPOINT ["go", "run", "./cmd/rfinterference"]
CMD ["--smoke-test"]
