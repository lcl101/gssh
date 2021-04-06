package main

//go mod tidy
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
)

var (
	v      = flag.Bool("v", false, "版本信息")
	help   = flag.Bool("help", false, "帮助")
	config = flag.String("c", "", "配置文件，默认al.conf")
	en     = flag.String("e", "", "加密密码")
	de     = flag.String("x", "", "licl")
	down   = flag.String("d", "", "下载配置文件")
	up     = flag.String("u", "", "下载配置文件")
)

func main() {
	cmdParse()
	version()

	downConfig()
	upConfig()

	upConfig()
	encrypt()
	defer func() {
		if err := recover(); err != nil {
			core.Log.Error("recover", err)
		}
	}()
	configFile := core.ReadConfigPath(*config)

	app := core.App{
		ConfigPath: configFile,
	}
	decrypt(&app)

	serverName := ""
	if len(os.Args) > 1 {
		serverName = os.Args[1]
	}
	core.Log.Info("登录服务器: ", serverName)
	app.Init(serverName)
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
		fmt.Println("gsh version: " + Version + ", Build " + Build + "。")
		fmt.Println("本程序源码：https://github.com/lcl101/rcmd。")
		os.Exit(0)
	}
}

func downConfig() {
	if *down != "" {
		//下载配置文件，暂时未实现
		fmt.Println("下载配置文件，功能稍后开放")
		os.Exit(0)
	}
}

func upConfig() {
	if *up != "" {
		//上传配置文件，暂时未实现
		fmt.Println("上传配置文件，功能稍后开放")
		os.Exit(0)
	}
}

func encrypt() {
	if *en != "" {
		fmt.Println(*en)
		s, err := core.Encrypt(*en)
		if err != nil {
			fmt.Println("en error: ", err)
		} else {
			fmt.Println(s)
		}
		os.Exit(0)
	}
}
func decrypt(app *core.App) {
	if *de != "" {
		fmt.Println("主机信息：", *de)
		d := app.ShowPasswd(*de)
		s, err := core.Decrypt(d)
		if err != nil {
			fmt.Println("de error: ", err)
		} else {
			fmt.Println(s)
		}
		os.Exit(0)
	}
}
