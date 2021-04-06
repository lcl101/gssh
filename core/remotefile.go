package core

import (
	"errors"
	"strings"

	"golang.org/x/crypto/ssh"
)

func RemoteIsExists(pathFile string, client *ssh.Client) (bool, error) {
	return remoteFileFun(pathFile, "e", client)
}

func RemoteIsFile(pathFile string, client *ssh.Client) (bool, error) {
	return remoteFileFun(pathFile, "f", client)
}

func remoteFileFun(pathFile, checkType string, client *ssh.Client) (bool, error) {
	if pathFile == "" {
		return false, errors.New("文件名为空")
	}
	cmd := NewCmd(client)
	//判断远程文件是否存在
	cmd.AddCmd("[ -" + checkType + " " + pathFile + " ] && echo 1 || echo 2")
	cmd.Run()
	if cmd.GetRtnCode() != 0 {
		Log.Error(cmd.ResultMsg())
		return false, errors.New("远程检查执行错误")
	}
	if trim(cmd.GetRtnMsg()) == "2" {
		return false, nil
	}
	return true, nil
}

func RemoteParsePath(pathFile string, client *ssh.Client) (string, error) {
	if pathFile == "" {
		return "", errors.New("文件名为空")
	}
	if pathFile[0] == '/' {
		return pathFile, nil
	}

	cmd := NewCmd(client)
	if pathFile[0] == '.' || pathFile[0] == '~' {
		cmd.AddCmd("echo $PWD" + pathFile[1:])
	} else {
		cmd.AddCmd("echo $PWD/" + pathFile)
	}
	cmd.Run()
	if cmd.GetRtnCode() != 0 {
		Log.Error(cmd.ResultMsg())
		return "", errors.New("远程检查执行错误")
	}

	return trim(cmd.rtnMsg), nil
}

func trim(str string) string {
	s := strings.Replace(str, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Trim(s, " ")
	return s
}
