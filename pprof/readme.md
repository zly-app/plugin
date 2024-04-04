
# 示例

```go
package main

import (
	"github.com/zly-app/plugin/pprof"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		pprof.WithPlugin(),
	)
	defer app.Exit()
	app.Run()
}
```

# 访问

```sh
curl 'http://localhost:6060/debug/pprof/'
curl 'http://localhost:6060/debug/pprof/profile?seconds=30'     # 默认进行 30s 的 CPU Profiling，得到一个分析用的 profile 文件
curl 'http://localhost:6060/debug/pprof/heap'        # 得到一个 heap 文件
```

# 可视化

windows在[这里](https://graphviz.gitlab.io/_pages/Download/Download_windows.html)下载并安装graphviz

ubuntu执行`apt install graphviz`安装graphviz
注意设置环境变量 path, 如果不能运行需要给 dot.exe 添加管理员执行权限

启动可视化工具渲染

```
go tool pprof -http=:80 profile
go tool pprof -http=:80 heap

go tool pprof -http=:80 http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof -http=:80 http://localhost:6060/debug/pprof/heap
```

# 配置文件

添加配置文件 `configs/default.yml`. 更多配置说明参考[这里](./config.go)

```yaml
plugins:
  pprof:
    Bind: :6060
    Enable: true
```
