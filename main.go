package main

import (
	"github.com/fireinrain/javbus-api/api"
	"github.com/fireinrain/javbus-api/config"
)

func main() {
	var conf = config.GlobalConfig
	// 2. 启动 API 服务
	api.RunApiServer(conf)
}
