package main

import (
	"flag"
	"fmt"
	"gssh/core"
	"os"
)

var (
	//Version 版本信息
	Version = "0.1.1"
	//Build 编译时间
	Build = "20190301"

	v      = flag.Bool("v", false, "版本信息")
	help   = flag.Bool("help", false, "帮助")
	config = flag.String("c", "", "配置文件，默认al.conf")
)

func main() {
	cmdParse()
	version()

	defer func() {
		if err := recover(); err != nil {
			core.Log.Error("recover", err)
		}
	}()

	serverName, codes := parseCmd()

	configFile := core.ReadConfigPath(*config)
	app := core.App{
		ConfigPath: configFile,
	}
	server, err := app.GetServer(serverName)
	if err != nil {
		core.Errorln("获取服务器错误！", err)
		os.Exit(1)
	}

	client, err := server.GenClient()
	if err != nil {
		core.Errorln("获取服务器连接错误!", err)
		os.Exit(1)
	}
	defer client.Close()

	cmd := core.NewCmd(client)
	cmd.SetCmds(codes)
	cmd.Run()
	if cmd.GetRtnCode() == 0 {
		core.Infoln(cmd.GetRtnMsg())
		return
	}
	core.Errorln("===========================")
	core.Errorln("执行命令异常:", cmd.ResultMsg())
	core.Errorln("===========================")

	// fmt.Println(app)
}

func parseCmd() (string, []string) {
	cmds := flag.Args()
	if len(cmds) < 2 {
		flag.Usage()
		core.Infoln("gcmd serverName 'ls -lart'")
		os.Exit(0)
	}
	return cmds[0], cmds[1:]
}

func cmdParse() {
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
}

func version() {
	if *v {
		fmt.Println("gcp version: " + Version + ", Build " + Build + "。")
		fmt.Println("本程序源码：https://github.com/lcl101/rcmd。")
		os.Exit(0)
	}
}
