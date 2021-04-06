package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/crypto/ssh"
)

//ErrorAssert 错误断言
func ErrorAssert(err error, assert string) bool {
	return strings.Contains(err.Error(), assert)
}

//Clear 清屏
func Clear() {
	var cmd exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = *exec.Command("cmd", "/c", "cls")
	} else {
		cmd = *exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

//ZhLen 计算字符宽度（中文）
func ZhLen(str string) int {
	length := 0
	for _, c := range str {
		if unicode.Is(unicode.Scripts["Han"], c) {
			length += 2
		} else {
			length++
		}
	}
	return length
}

//ParseAuthMethods ssh解析鉴权方式
func ParseAuthMethods(passwd, key string) ([]ssh.AuthMethod, error) {
	sshs := []ssh.AuthMethod{}

	if passwd != "" {
		sshs = append(sshs, ssh.Password(passwd))
		return sshs, nil
	}
	method, err := pemKey(key)
	if err != nil {
		return nil, err
	}
	sshs = append(sshs, method)
	return sshs, nil
}

// 解析密钥
func pemKey(key string) (ssh.AuthMethod, error) {
	sshKey := key
	if sshKey == "" {
		sshKey = "~/.ssh/id_rsa"
	}
	sshKey, _ = ParsePath(sshKey)

	pemBytes, err := os.ReadFile(sshKey)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

//GetExecPath 获取当前路径
func GetExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\"`)
	}
	return string(path[0 : i+1]), nil
}
