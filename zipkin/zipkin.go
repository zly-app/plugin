package zipkin

import (
	"fmt"
	"net"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkin_model "github.com/openzipkin/zipkin-go/model"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/zly-app/zapp/core"
)

type ZipKinPlugin struct {
	app core.IApp
}

func NewZipKinPlugin(app core.IApp) core.IPlugin {
	return &ZipKinPlugin{app}
}

func (z *ZipKinPlugin) Inject(a ...interface{}) {}

func (z *ZipKinPlugin) Start() error {
	var conf Config
	key := "plugins." + string(nowPluginType)

	vi := z.app.GetConfig().GetViper()
	if vi.IsSet(key) {
		err := vi.UnmarshalKey(key, &conf)
		if err != nil {
			return fmt.Errorf("无法解析<%s>插件配置: %s", nowPluginType, err)
		}
	}

	if conf.ApiUrl == "" {
		conf.ApiUrl = defaultApiUrl
	}
	if conf.IP == "" {
		conf.IP = defaultIP
	}
	return z.InitTracer(&conf)
}

func (z *ZipKinPlugin) Close() error {
	return nil
}

func (z *ZipKinPlugin) InitTracer(conf *Config) error {
	reporter := zipkinhttp.NewReporter(conf.ApiUrl)
	endpoint := &zipkin_model.Endpoint{
		ServiceName: z.app.Name(),
		IPv4:        net.ParseIP(conf.IP),
		Port:        conf.Port,
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return fmt.Errorf("无法创建zipkin跟踪程序: %+v\n", err)
	}

	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)
	return nil
}
