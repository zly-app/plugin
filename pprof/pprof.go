package pprof

import (
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/zly-app/zapp/core"
	"go.uber.org/zap"
)

func NewPProf(app core.IApp) core.IPlugin {
	conf := newConfig()
	err := app.GetConfig().ParsePluginConfig(defPluginType, conf, true)
	if err != nil {
		app.Fatal("pprof配置错误", zap.Error(err))
	}
	conf.Check()

	p := &PProf{
		conf: conf,
	}
	return p
}

type PProf struct {
	conf *Config
	mux  *http.ServeMux
	srv  *http.Server
}

func (p *PProf) Inject(a ...interface{}) {}

func (p *PProf) Start() error {
	if p.conf.Disable {
		return nil
	}

	p.mux = http.NewServeMux()
	p.mux.HandleFunc("/debug/pprof/", pprof.Index)
	p.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	p.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	p.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	p.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	l, err := net.Listen("tcp", p.conf.Bind)
	if err != nil {
		return err
	}

	srv := &http.Server{Handler: p.mux}
	go srv.Serve(l)
	return nil
}

func (p *PProf) Close() error {
	if p.srv != nil {
		return p.srv.Close()
	}
	return nil
}
