# RF Interference Attribution

无线干扰事件归因服务：接收多站扫描片段，校正校准时钟，将相容证据归并为干扰事件并保存可解释的归因快照。

```bash
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go run ./cmd/rfinterference --addr :8080
curl -X POST http://localhost:8080/v1/demo/import
curl http://localhost:8080/v1/events
```

SQLite 数据库默认保存在 `rfinterference.db`，可用 `--db` 修改。`--smoke-test` 会使用临时数据库运行自检后退出。
启动服务后访问 `http://localhost:8080/` 可打开最小操作页，触发示例导入和自检。
