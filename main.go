package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 创建Gin服务器
	addr := fmt.Sprintf(":%d", conf.Server.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: api.SetupRouter(conf), // 修改RunApiServer为SetupRouter返回路由器
	}
	// 启动服务
	log.Printf("Server starting on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("Server failed: %v", err))
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("关闭服务器...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}
	//关闭db 关闭cache 关闭定时任务

}
