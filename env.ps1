# 在 trade-hub 目录执行: . .\env.ps1
# 国内下载 Go 依赖用（解决 missing go.sum / proxy.golang.org 超时）
$env:GOPROXY = "https://goproxy.cn,direct"
