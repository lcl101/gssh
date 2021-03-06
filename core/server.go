package core

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

//Server 定义服务器
type Server struct {
	Name     string                 `json:"name"`
	IP       string                 `json:"ip"`
	Port     int                    `json:"port"`
	User     string                 `json:"user"`
	Password string                 `json:"password"`
	Method   string                 `json:"method"`
	Key      string                 `json:"key"`
	Options  map[string]interface{} `json:"options"`

	termWidth  int
	termHeight int
}

//Format 格式化
func (server *Server) Format() {
	if server.Port == 0 {
		server.Port = 22
	}

	if server.Method == "" {
		server.Method = "password"
	}
}

func (server *Server) GenClient() (*ssh.Client, error) {
	pw := server.Password
	key := server.Key
	if server.Method == "k" {
		pw = ""
	} else {
		if pw != "" {
			passwd, err := Decrypt(pw)
			if err != nil {
				Errorln("密码解析错误:", err)
				Log.Error("密码解析错误:", err)
				passwd = server.Password
			}
			pw = passwd
		}
	}
	auths, err := ParseAuthMethods(pw, key)

	if err != nil {
		Errorln("鉴权出错:", err)
		Log.Error("auth fail", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 默认端口为22
	if server.Port == 0 {
		server.Port = 22
	}

	addr := server.IP + ":" + strconv.Itoa(server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		if ErrorAssert(err, "ssh: unable to authenticate") {
			Errorln("连接失败，请检查密码/密钥是否有误")
			return nil, err
		}
		Errorln("ssh dial fail:", err)
		Log.Error("ssh dial fail", err)
		return nil, err
	}
	return client, nil
}

//Connect 执行远程连接
func (server *Server) Connect() {
	client, err := server.GenClient()
	if err != nil {
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		Errorln("create session fail:", err)
		Log.Error("create session fail", err)
		return
	}

	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		Errorln("创建文件描述符出错:", err)
		Log.Error("创建文件描述符出错", err)
		return
	}

	stopKeepAliveLoop := server.startKeepAliveLoop(session)
	defer close(stopKeepAliveLoop)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	defer term.Restore(fd, oldState)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	server.termWidth, server.termHeight, _ = term.GetSize(fd)
	if err := session.RequestPty("xterm-256color", server.termHeight, server.termWidth, modes); err != nil {
		Errorln("创建终端出错:", err)
		Log.Error("创建终端出错", err)
		return
	}

	winChange := server.listenWindowChange(session, fd)
	defer close(winChange)

	err = session.Shell()
	if err != nil {
		Errorln("执行Shell出错:", err)
		Log.Error("执行Shell出错", err)
		return
	}

	err = session.Wait()
	if err != nil {
		//Errorln("执行Wait出错:", err)
		Log.Error("执行Wait出错", err)
		return
	}
}

// 监听终端窗口变化
func (server *Server) listenWindowChange(session *ssh.Session, fd int) chan struct{} {
	terminate := make(chan struct{})
	go func() {
		for {
			select {
			case <-terminate:
				return
			default:
				termWidth, termHeight, _ := term.GetSize(fd)

				if server.termWidth != termWidth || server.termHeight != termHeight {
					server.termHeight = termHeight
					server.termWidth = termWidth
					session.WindowChange(termHeight, termWidth)
				}

				time.Sleep(time.Millisecond * 3)
			}
		}
	}()

	return terminate
}

// 发送心跳包
func (server *Server) startKeepAliveLoop(session *ssh.Session) chan struct{} {
	terminate := make(chan struct{})
	go func() {
		for {
			select {
			case <-terminate:
				return
			default:
				if val, ok := server.Options["ServerAliveInterval"]; ok && val != nil {
					_, err := session.SendRequest("kl@licl", true, nil)
					if err != nil {
						Log.Error("keepAliveLoop fail", err)
					}
					// Log.Info("kl....")
					t := time.Duration(server.Options["ServerAliveInterval"].(float64))
					time.Sleep(time.Second * t)
				}
			}
		}
	}()
	return terminate
}

//MergeOptions 合并选项
func (server *Server) MergeOptions(options map[string]interface{}, overwrite bool) {
	if server.Options == nil {
		server.Options = make(map[string]interface{})
	}

	for k, v := range options {
		if overwrite {
			server.Options[k] = v
		} else {
			if _, ok := server.Options[k]; !ok {
				server.Options[k] = v
			}
		}

	}
}

//Edit 编辑服务配置
func (server *Server) Edit() {
	input := ""
	Info("Name(default=" + server.Name + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Name = input
		input = ""
	}

	Info("Ip(default=" + server.IP + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.IP = input
		input = ""
	}

	Info("Port(default=" + strconv.Itoa(server.Port) + ")：")
	fmt.Scanln(&input)
	if input != "" {
		port, _ := strconv.Atoi(input)
		server.Port = port
		input = ""
	}

	Info("User(default=" + server.User + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.User = input
		input = ""
	}

	Info("Password(default=" + server.Password + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Password = input
		input = ""
	}

	Info("Method(default=" + server.Method + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Method = input
		input = ""
	}

	Info("Key(default=" + server.Key + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Key = input
		input = ""
	}
}
