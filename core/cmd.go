package core

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Cmd struct {
	client  *ssh.Client
	codes   []string
	rtnCode int
	rtnMsg  string
}

func NewCmd(client *ssh.Client) *Cmd {
	return &Cmd{
		client:  client,
		codes:   []string{},
		rtnCode: -1,
		rtnMsg:  "",
	}
}

func (c *Cmd) SetCmds(codes []string) {
	c.codes = codes
}

func (c *Cmd) AddCmd(code string) {
	if code == "" {
		return
	}
	c.codes = append(c.codes, code)
}

func (c *Cmd) GetRtnCode() int {
	return c.rtnCode
}

func (c *Cmd) GetRtnMsg() string {
	return c.rtnMsg
}

func (c *Cmd) ResultMsg() string {
	if c.rtnCode == 0 {
		return c.rtnMsg
	}
	return fmt.Sprintf("rtnCode=[%d],rtnMsg=[%s]", c.rtnCode, c.rtnMsg)
}

func (c *Cmd) Run() {
	session, err := c.client.NewSession()

	if err != nil {
		Errorln("create session fail:", err)
		Log.Error("create session fail", err)
		c.rtnCode = 12
		c.rtnMsg = "create session fail!"
		return
	}
	defer session.Close()
	codes := append(c.codes, "exit")
	cmd := strings.Join(codes, "&&")

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	Log.Info("run cmd : ", cmd)

	err = session.Run(cmd)
	if err != nil {
		Log.Error("run cmd fail", err)
		c.rtnCode = 10
		c.rtnMsg = err.Error()
		return
	}

	if stderr.String() != "" {
		c.rtnCode = 11
		c.rtnMsg = stderr.String()
		return
	}

	c.rtnCode = 0
	c.rtnMsg = stdout.String()
}
