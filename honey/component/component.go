package component

import (
	"fmt"

	"github.com/zly-app/zapp/core"
)

type IComponent interface {
	core.IComponent
	/*解析上报者配置数据到结构中
	  配置项在配置文件中为 [reporter.{输入设备名}]
	  name 上报者名
	  outPtr 接收配置的变量
	  ignoreNotSet 如果无配置数据, 则忽略, 默认为false
	*/
	ParseReporterConf(name string, conf interface{}, ignoreNotSet ...bool) error
}

var _ IComponent = (*Component)(nil)

type Component struct {
	core.IComponent
}

// 解析上报者配置数据到结构中
func (c *Component) ParseReporterConf(name string, conf interface{}, ignoreNotSet ...bool) error {
	key := fmt.Sprintf("plugins.honey.%s", name)
	return c.App().GetConfig().Parse(key, conf, ignoreNotSet...)
}

// 获取Component
func NewComponent(c core.IComponent) IComponent {
	return &Component{
		IComponent: c,
	}
}
