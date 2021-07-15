package zipkin

const (
	// 默认api地址
	defaultApiUrl = "http://localhost:9411/api/v2/spans"
	// 默认ip
	defaultIP = "127.0.0.1"
)

type Config struct {
	ApiUrl string // api地址, 默认为 http://localhost:9411/api/v2/spans
	IP     string // 对外展示ip, 用于区分上报信息来自哪儿
	Port   uint16 // 对外展示端口, 用于区分上报信息来自哪儿
}
