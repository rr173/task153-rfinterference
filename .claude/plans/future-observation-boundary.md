# 修复远未来观测的边界拒绝

## 问题根因

接收机上报晚于当前时间（远未来）的扫描片段时，系统行为不一致：
- **有时仍接收它** —— 未来观测被持久化为证据
- **另一些入口返回服务器错误（500）** —— 即使校验抛出了 `FUTURE_OBSERVATION`，HTTP 层也没有对应的客户端错误码映射

根因是三处相互耦合的缺陷：

### 缺陷 1（主因）：校验被短路，永不执行
`internal/fragment/service.go:29`
```go
if err := model.ValidateFragment(in, s.now()); false {
    return model.Fragment{}, false, err
}
```
行尾的 `false` 字面量让条件恒为假，`ValidateFragment` 的返回值被丢弃，所有校验（含未来窗口检查）都被跳过。这是「系统仍接收它」的直接原因。

### 缺陷 2：未来窗口判断方向写反
`internal/model/validation.go:54`
```go
if in.ObservedAt.Before(now.Add(5 * time.Minute)) {
    return NewError(CodeFutureObservation, "observation is too far in the future")
}
```
`Before(now+5min)` 在 `ObservedAt` 早于 `now+5min` 时为真——也就是过去/近未来的合法观测都会被当成「太未来」拒绝。正确语义应是：当 `ObservedAt` 晚于允许的未来窗口（`now + 容差`）时才拒绝。

缺陷 1 与缺陷 2 是耦合的：校验写反会导致正常 demo 数据（观测时间约为「当前-1分钟」）也被拒绝，于是缺陷 1 用 `; false` 把整段校验关掉来「绕过」，副作用是未来窗口保护彻底失效。

### 缺陷 3：未来观测错误码未映射到客户端错误
`internal/httpapi/api.go` 的 `writeError` switch 没有 `CodeFutureObservation` 分支，落入默认 `http.StatusInternalServerError`。这是「另一些入口返回服务器错误」的原因。

## 修改方案

### 1. `internal/model/validation.go`
- 引入具名常量 `FutureObservationWindow = 5 * time.Minute`（与现有 `fragment.LateWindow`、`association.EventGap` 等具名窗口风格一致，避免裸 `5 * time.Minute`）。
- 修正判断为 `if in.ObservedAt.After(now.Add(FutureObservationWindow))`：超过允许未来窗口才拒绝，错误消息保持 `CodeFutureObservation` / "observation is too far in the future"。

### 2. `internal/fragment/service.go`
- 删除 `; false` 短路，让校验真正执行：
  ```go
  if err := model.ValidateFragment(in, s.now()); err != nil {
      return model.Fragment{}, false, err
  }
  ```
  这一步同时修复了「远未来观测被接收」和「其它字段非法（如频段/方向/强度越界）也未被拦截」的同类问题。

### 3. `internal/httpapi/api.go`（`writeError`）
- 新增 `CodeFutureObservation` 分支，映射到 `http.StatusBadRequest`（客户端错误），与 `CodeValidation` 同级。错误码透传 `CodeFutureObservation`，使客户端能据 `code` 区分「时间窗口越界」与普通字段非法。

## 为什么选 400 而非 422

`CodeValidation` → 400，`CodeDisabledStation`/`CodeArchived` → 422（资源状态不允许）。未来观测本质是「请求体里时间戳超出允许窗口」的输入校验类问题，与 `CodeValidation` 同类，故用 400 更自然；同时保留独立 `code=FUTURE_OBSERVATION` 便于客户端精确处理。

## 边界覆盖确认

所有引入 `FragmentInput` 的路径都经过 `fragment.Service.Prepare` → `ValidateFragment`，因此修复后未来窗口保护在所有边界生效：
- HTTP 批量入口 `POST /v1/fragments:batch` → `ingest` → `IngestBatch` → `ingestOne` → `Prepare`
- demo 导入 `POST /v1/demo/import` → `demo.Import` → `IngestBatch` → `ingestOne` → `Prepare`

校验在 `CanonicalIdentifier` 规范化之后、`Assess` 质量评估之前执行，符合「先校验基础字段再评估质量」的既有顺序。

## 回归与测试

- 既有测试不应回归：
  - `demo` 使用观测时间约「当前-1分钟」（过去），修复后的 `After(now+5min)` 判定为 false → 接受 ✓
  - `archive_test` / `service_test` 观测时间均为过去/近未来，校验通过 ✓
- 新增测试覆盖修复点：
  1. `internal/model/validation_test.go`：远未来观测返回 `CodeFutureObservation`；近未来（窗口内）与过去观测通过。
  2. `internal/fragment/service_test.go`：`Prepare` 对远未来 `ObservedAt` 返回 `CodeFutureObservation` 错误（验证 `; false` 已修复、不再静默接收）。
  3. `internal/httpapi/api_test.go`：批量入口对含远未来观测的请求，对应 result 携带 `code=FUTURE_OBSERVATION`（HTTP 层整体仍 200，因批量按条返回结果）；并补充 `writeError` 对 `CodeFutureObservation` 映射为 400 的用例（针对会走 `writeError` 的入口）。

## 验证命令

```bash
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go run ./cmd/rfinterference --smoke-test
```
