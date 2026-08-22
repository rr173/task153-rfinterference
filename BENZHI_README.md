# 构建与运行

本项目提供多接收站扫描片段的持久化、事件关联和归因 API。
启动 HTTP 服务后访问根路径可使用面向值班工程师的最小操作页，触发示例导入和自检。

```bash
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go run ./cmd/rfinterference --addr :8080
GOTOOLCHAIN=local go run ./cmd/rfinterference --smoke-test
```

构建 Docker 镜像：`./build_benzhi_docker.sh my-rf-service linux/amd64`。容器镜像以 shell 为入口，可执行 `go run ./cmd/rfinterference --smoke-test` 进行无外部依赖自检。
