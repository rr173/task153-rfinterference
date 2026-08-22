package httpapi

import "net/http"

const operatorPage = `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>无线干扰事件归因</title></head>
<body><main><h1>无线干扰事件归因</h1><p>查看服务健康状态、导入示例证据并查询当前事件。</p><p><button id="check">运行自检</button> <button id="demo">导入示例</button></p><pre id="output">等待操作</pre><script>
const output = document.getElementById('output');
async function call(path, method) { const response = await fetch(path, {method}); output.textContent = await response.text(); }
document.getElementById('check').onclick = () => call('/v1/self-check', 'GET');
document.getElementById('demo').onclick = () => call('/v1/demo/import', 'POST');
</script></main></body></html>`

func (a *API) page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(operatorPage))
}
