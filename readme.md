
# 插件

> 提供用于 https://github.com/zly-app/zapp 的插件

# 说明

> 插件的使用基本按照以下顺序

1. 启用插件, 在 zapp.NewApp 提供选项, 一般为 plugin.WithPlugin()
2. 注入插件, 这一步根据不同的插件有不同的方法, 一般为 plugin.RegistryXXX(...)
3. 启动app
