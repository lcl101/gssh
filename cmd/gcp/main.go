package main

import (
	"flag"
	"fmt"
	"gssh/core"
	"gssh/core/scp"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
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

	configFile := core.ReadConfigPath(*config)
	app := core.App{
		ConfigPath: configFile,
	}

	src, dest := parsePath(&app)
	var client *ssh.Client
	var err error
	if src.IsRemote() {
		client, err = src.GetClient()
		if err != nil {
			core.Errorln("获取ssh client错误", err)
		}
	}
	if dest.IsRemote() {
		client, err = dest.GetClient()
		if err != nil {
			core.Errorln("获取ssh client错误", err)
		}
	}

	s := scp.NewSCP(client)

	if src.IsRemote() {
		if src.IsDir() {
			s.ReceiveDir(src.PathFile(), dest.PathFile(), nil)
			return
		}
		s.ReceiveFile(src.PathFile(), dest.PathFile())
		return
	}
	if src.IsDir() {
		s.SendDir(src.PathFile(), dest.PathFile(), nil)
		return
	}

	s.SendFile(src.PathFile(), dest.PathFile())
	// fmt.Println(app)
}

func parsePath(app *core.App) (*GcpPath, *GcpPath) {
	srcPath := strings.TrimRight(flag.Arg(0), " ")
	descPath := strings.TrimRight(flag.Arg(1), " ")

	if srcPath == "" || descPath == "" {
		flag.Usage()
		core.Infoln("gcp 源文件 目标文件")
		os.Exit(0)
	}
	src, err := newGcpPath(srcPath, SRC_PATH, app)
	dest, err := newGcpPath(descPath, DEST_PATH, app)
	if err != nil {
		core.Errorln(err)
		os.Exit(0)
	}
	if err != nil {
		core.Errorln(err)
		os.Exit(0)
	}

	if src.IsDir() && !dest.IsDir() {
		core.Errorln("源是目录，目标是一个文件，请检查")
		os.Exit(0)
	}

	core.Infoln("--------------------------------------------")
	core.Info(fmt.Sprintf("源文件: [%s, %s]", src.path, src.fileName))
	core.Info("  ====>   ")
	core.Infoln(fmt.Sprintf("目标文件: [%s, %s]", dest.path, dest.fileName))
	core.Infoln("--------------------------------------------")
	return src, dest
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
