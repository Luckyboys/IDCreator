package main

import (
	"flag"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/Luckyboys/IDCreator/Service"
)

func main() {

	configPath := flag.String("c", "", "请加上配置文件")
	flag.Parse()

	Common.GetConfigInstance().Init(*configPath)
	Common.GetConfigInstance().PrintData()

	Service.StartToListen()

}
