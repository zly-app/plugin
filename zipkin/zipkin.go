package zipkin

import (
	"net"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkin_model "github.com/openzipkin/zipkin-go/model"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
)

type ZipKinPlugin struct {
	app core.IApp
}

func NewZipKinPlugin(app core.IApp) core.IPlugin {
	var conf Config

	// 解析配置
	key := "plugins." + string(nowPluginType)
	vi := app.GetConfig().GetViper()
	if vi.IsSet(key) {
		err := vi.UnmarshalKey(key, &conf)
		if err != nil {
			app.Fatal("无法解析插件配置", zap.String("PluginType", string(nowPluginType)), zap.Error(err))
		}
	}

	if conf.ApiUrl == "" {
		conf.ApiUrl = defaultApiUrl
	}
	if conf.IP == "" {
		conf.IP = defaultIP
	}

	reporter := zipkinhttp.NewReporter(conf.ApiUrl)
	endpoint := &zipkin_model.Endpoint{
		ServiceName: app.Name(),
		IPv4:        net.ParseIP(conf.IP),
		Port:        conf.Port,
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		app.Fatal("无法创建zipkin跟踪程序", zap.Error(err))
	}
	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)

	return &ZipKinPlugin{app}
}

func (z *ZipKinPlugin) Inject(a ...interface{}) {}

func (z *ZipKinPlugin) Start() error { return nil }

func (z *ZipKinPlugin) Close() error {
	return nil
}
