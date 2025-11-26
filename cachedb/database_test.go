package cachedb

import (
	"testing"

	"github.com/fireinrain/javbus-api/config"
)

func TestInitDataBase(t *testing.T) {
	initConfig, err := config.InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.DatabaseConfig = initConfig.DATABASE

	CacheDb, err = InitDataBase(cfg)
	//检查是否初始化成功
	if err != nil {
		t.Errorf("InitDataBase() error = %v", err)
	}
}
