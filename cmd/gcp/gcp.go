package main

import (
	"errors"
	"gssh/core"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	LOCAL     = "LOCAL"
	SRC_PATH  = 1
	DEST_PATH = 2
)

type GcpPath struct {
	app        *core.App
	serverName string
	path       string
	fileName   string
	pathType   int //src:1, dest:2
}

func newGcpPath(path string, pathType int, app *core.App) (*GcpPath, error) {
	serverName := LOCAL
	pathFile := path
	i := strings.Index(path, ":")
	if i > 0 {
		serverName = path[0:i]
		pathFile = path[i+1:]
	}
	if pathFile == "" {
		pathFile = "./"
	}
	if serverName == "" {
		return nil, errors.New("serverName is nil")
	}

	gp := &GcpPath{
		app:        app,
		serverName: serverName,
		path:       pathFile,
		fileName:   "",
		pathType:   pathType,
	}
	err := gp.init()
	return gp, err
}

func (gcp *GcpPath) IsDir() bool {
	return gcp.fileName == ""
}

func (gcp *GcpPath) PathFile() string {
	if gcp.fileName == "" {
		return gcp.path
	}
	return filepath.Join(gcp.path, gcp.fileName)
}

func (gcp *GcpPath) IsRemote() bool {
	return gcp.serverName != LOCAL
}

func (gcp *GcpPath) GetClient() (*ssh.Client, error) {
	if gcp.serverName == LOCAL {
		return nil, errors.New("local path")
	}
	server, err := gcp.app.GetServer(gcp.serverName)
	if err != nil {
		core.Errorln("获取服务器错误！", err)
		return nil, err
	}
	client, err := server.GenClient()
	if err != nil {
		core.Errorln("获取服务器连接错误!", err)
		return nil, err
	}
	return client, nil
}

func (gcp *GcpPath) init() error {
	if gcp.serverName == LOCAL {
		return gcp.local()
	}
	return gcp.remote()
}

func (gcp *GcpPath) remote() error {
	// server, err := gcp.app.GetServer(gcp.serverName)
	// if err != nil {
	// 	core.Errorln("获取服务器错误！", err)
	// 	return err
	// }
	// client, err := server.GenClient()
	// if err != nil {
	// 	core.Errorln("获取服务器连接错误!", err)
	// 	return err
	// }
	client, err := gcp.GetClient()
	if err != nil {
		core.Errorln("获取ssh client错误！", err)
		return err
	}
	defer client.Close()
	path, err := core.RemoteParsePath(gcp.path, client)
	if err != nil {
		core.Log.Error("执行远程命令错误", err)
		return err
	}

	if gcp.pathType == SRC_PATH {
		//如果是源文件，需要判断远程文件是否存在
		rb, err := core.RemoteIsExists(path, client)
		if err != nil {
			core.Log.Error("读取远程文件信息错误", err)
			return err
		}
		if !rb {
			return errors.New("源远程文件不存在：" + gcp.path)
		}
		//如果是源文件，需要判断是否是文件，而不是文件夹
		rb, err = core.RemoteIsFile(path, client)
		if err != nil {
			core.Log.Error("读取远程文件信息错误", err)
			return err
		}
		if rb {
			// return errors.New("源远程不是文件：" + gcp.path)
			gcp.path = filepath.Dir(path)
			gcp.fileName = filepath.Base(path)
			return nil
		}
		gcp.path = filepath.Dir(path)
		gcp.fileName = ""
		return nil
	}
	//如果是目标文件，判断逻辑：
	//需要判断是否存在
	//	如果存在判断是文件，说明文件存在，报错
	//	如果是文件夹，则正确
	//如果不存在
	//	判断上层目录是否存在
	rfe, err := core.RemoteIsExists(path, client)
	if err != nil {
		core.Log.Error("读取远程文件信息错误", err)
		return err
	}
	if rfe {
		gcp.path = path
		gcp.fileName = ""
		rf, err := core.RemoteIsFile(path, client)
		if err != nil {
			core.Log.Error("读取远程文件信息错误", err)
			return err
		}
		if rf {
			//目标文件已经存在了，目前采用报错机制
			return errors.New("目标文件已经存在")
		}
		return nil
	}
	//远程文件不存在
	gcp.path = filepath.Dir(path)
	gcp.fileName = filepath.Base(path)
	//判断上层文件夹是否存在
	rb, err := core.RemoteIsExists(gcp.path, client)
	if err != nil {
		core.Log.Error("读取远程文件信息错误", err)
		return err
	}
	if !rb {
		return errors.New("目标文件夹不存在")
	}
	return nil
}

func (gcp *GcpPath) local() error {
	path, err := core.ParsePath(gcp.path)
	if err != nil {
		core.Log.Error(err)
		return err
	}
	if gcp.pathType == SRC_PATH {
		if !core.IsExist(path) {
			return errors.New("源本地文件不存在：" + gcp.path)
		}
		if core.IsFile(path) {
			gcp.path = core.PathName(path)
			gcp.fileName = core.FileName(path)
			return nil
		}
		gcp.path = core.PathName(path)
		gcp.fileName = ""
		return nil
	}
	//如果是目标文件
	if core.IsExist(path) {
		gcp.path = path
		gcp.fileName = ""
		if core.IsFile(path) {
			//目标文件已经存在了，目前采用报错机制
			return errors.New("目标文件已经存在")
		}
		return nil
	}
	//目标文件不存在
	gcp.path = core.PathName(path)
	gcp.fileName = core.FileName(path)
	if !core.IsExist(gcp.path) {
		return errors.New("目标文件夹不存在")
	}
	return nil
}
