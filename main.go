package main

import (
	"fmt"

	"github.com/fireinrain/javbus-api/api"
	"github.com/fireinrain/javbus-api/cachedb"
	"github.com/fireinrain/javbus-api/config"
)

func main() {
	//初始化配置
	conf, err := config.InitConfig()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	//初始化db
	db, err := cachedb.InitDataBase(conf.DATABASE)
	if err != nil {
		panic(fmt.Sprintf("初始化数据库失败: %v", err))
	}
	_ = db
	//初始化dao层

	//启动定时任务

	// 启动 API 服务
	api.RunApiServer(conf)
}
