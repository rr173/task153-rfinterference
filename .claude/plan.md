# 修复跨越正北边界的方向比较、关联与归因

## 问题根因

方位角是环形的（0/360 边界处回绕），但代码在三处按**线性**数值处理方向，导致：

1. **`model.DirectionDeviation`**（`internal/model/evidence.go:62`）：`delta = |a - b|`，对 359° vs 1° 返回 358 而非 2。直接污染 `DirectionCompatible` 与 `directionScore`。
2. **`association.directionRange`**（`internal/attribution/compute.go:64`）：对方向排序后返回 `last - first` 作为跨度。对 `[359, 1]` 排序后为 `[1, 359]`，跨度=358 > 90 → 判定 `VerdictDirectionConflict`。这正是任务描述的"被判为方向冲突、无法形成可定位事件"。同时 `DirectionMin/DirectionMax` 被写成 1..359，丢失了正北附近的连续含义。
3. **`association.DirectionCompatible`**（`internal/association/rules.go:26`）：判定条件写反——`if distance >= tolerance { return true }`，即"方向离得远才兼容"。同向证据（20° vs 27°，差 7°）被误判为冲突而无法合并，已有 `TestCompatibilityRules` 与 `TestIngestGroupsThreeStations` 验证这一点。应当是"存在一个旧方向与新方向容差内"才兼容（`<=`）。

冒烟自检能过是因为 `SelfCheck` 只看 `>=1 event && >=3 fragments`（`CountFragments` 计入被排除片段），掩盖了关联失败。

## 修改方案

### 1. `internal/model/evidence.go` — 环形偏差
```go
func DirectionDeviation(a, b float64) float64 {
	delta := math.Abs(NormalizeDirection(a) - NormalizeDirection(b))
	if delta > 180 {
		delta = 360 - delta
	}
	return delta
}
```
（`NormalizeDirection` 已保证输入落入 `[0,360)`，故线性差 ∈ `[0,360)`；`>180` 时取补角。）

### 2. `internal/association/rules.go` — 修复反转的比较
```go
func DirectionCompatible(existing []model.Fragment, f model.Fragment) bool {
	if len(existing) == 0 {
		return true
	}
	for _, old := range existing {
		if DirectionDistance(old.DirectionDeg, f.DirectionDeg) <= DirectionToleranceDeg {
			return true
		}
	}
	return false
}
```
`DirectionDistance` 已是环形的（修复后），`directionScore` 语义保持不变（越近分越高）。

### 3. `internal/attribution/compute.go` — 环形跨度与端点
```go
func directionRange(values []float64) (float64, float64, float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	norm := make([]float64, len(values))
	for i, v := range values {
		norm[i] = model.NormalizeDirection(v)
	}
	sort.Float64s(norm)
	if len(norm) == 1 {
		return norm[0], norm[0], 0
	}
	// 最大相邻间隙（含首尾环回），弧 = 360 - 最大间隙
	maxGap := 0.0
	maxIdx := 0
	for i := 0; i < len(norm); i++ {
		next := norm[(i+1)%len(norm)]
		gap := next - norm[i]
		if gap < 0 {
			gap += 360
		}
		if gap > maxGap {
			maxGap, maxIdx = gap, i
		}
	}
	span := 360 - maxGap
	min := norm[(maxIdx+1)%len(norm)]
	max := norm[maxIdx]
	return min, max, span
}
```
对 `[359, 1]`：排序 `[1, 359]`，间隙 = `358`（359→1 环回）与 `2`（1→359），maxGap=358，span=2，min=`norm[1]`=359，max=`norm[0]`=1 → 报告 `359..1` 跨越正北、跨度 2。对 `[20,28,35]`：maxGap 为端点环回间隙 345，span=15，min=20,max=35。

`Compute` 中的 `span > 90` 阈值与 `BuildConfidence` 的 `directionSpan/90` 无需改动，输入的 span 现在是环形值。

### 4. 测试
- 新增 `internal/model/evidence_test.go`：验证 `DirectionDeviation(359,1)=2`、`(1,359)=2`、`(350,10)=20`、`(180,170)=10`、`(0,0)=0`。
- 扩充 `internal/association/rules_test.go`：保留原断言（200° vs 10° 不兼容，修复反转后也仍 false，因环形差=170>45），新增 359° vs 1° 兼容、20° vs 27° 兼容的断言。
- 新增 `internal/attribution/compute_test.go` 跨越正北用例：359/1 两站 → `VerdictLocatable`（或补足≥2站与可信校准）且 `DirectionMin/Max` 跨越正北、span≈2。

## 影响面与一致性
- `directionScore`、`DirectionDistance`、`BuildConfidence` 行为与环形偏差自洽，无需改动。
- `DirectionMin/DirectionMax` 现在可能 `min > max`（跨越 0/360），这是环形弧的有意表达，`store/snapshots.go` 仅作存储与回放，不依赖数值大小关系。
- 不触碰 DB schema、验证范围 `[0,360)`、归一化与摄入路径。

## 验证
`GOTOOLCHAIN=local go test ./...` 与 `GOTOOLCHAIN=local go run ./cmd/rfinterference --smoke-test`，前者全绿，后者仍输出 1 event / 3 fragments。
