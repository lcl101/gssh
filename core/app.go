package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//IndexType index类型
type IndexType int

//IndexType 枚举类型
const (
	IndexTypeServer IndexType = iota
	IndexTypeGroup
)

const (
	//SP 显示名称的空格
	SP = "                   "
)

//Group 分组
type Group struct {
	GroupName string   `json:"group_name"`
	Prefix    string   `json:"prefix"`
	Servers   []Server `json:"servers"`
}

//Config 配置文件
type Config struct {
	ShowDetail bool                   `json:"show_detail"`
	Servers    []Server               `json:"servers"`
	Groups     []Group                `json:"groups"`
	Options    map[string]interface{} `json:"options"`
}

//ServerIndex 服务索引
type ServerIndex struct {
	indexType   IndexType
	groupIndex  int
	serverIndex int
	server      *Server
}

//App app结构
type App struct {
	ConfigPath  string
	config      Config
	serverIndex map[string]ServerIndex
}

func (app *App) GetServer(serverName string) (*Server, error) {
	if serverName == "" {
		return nil, errors.New("serverName is nil")
	}
	app.serverIndex = make(map[string]ServerIndex)
	// 解析配置
	app.loadConfig()
	app.loadServerMap(true)
	servers := app.config.Servers

	for _, v := range servers {
		if v.Name == serverName {
			return &v, nil
		}
	}
	return nil, errors.New("not found server")
}

//ShowPasswd 获取加密密码
func (app *App) ShowPasswd(serverName string) string {
	if serverName == "" {
		return ""
	}
	app.serverIndex = make(map[string]ServerIndex)
	// 解析配置
	app.loadConfig()
	app.loadServerMap(true)
	servers := app.config.Servers
	var s Server
	for _, v := range servers {
		if v.Name == serverName {
			s = v
			break
		}
	}

	if s.Password != "" {
		return s.Password
	}
	return serverName
}

//Init 执行脚本
func (app *App) Init(serverName string) {
	app.serverIndex = make(map[string]ServerIndex)

	// 解析配置
	app.loadConfig()

	app.loadServerMap(true)

	if serverName == "" {
		app.show()
	} else {
		servers := app.config.Servers
		var s Server
		for _, v := range servers {
			if v.Name == serverName {
				s = v
				break
			}
		}
		if s.Name != "" {
			app.TipsMsg(s.Name)
			s.Connect()
		} else {
			app.show()
		}
	}
}

func (app *App) TipsMsg(name string) {
	Infoln("================================")
	Info("您登录的服务器：")
	Errorln(name)
	Infoln("================================")
}

func (app *App) saveAndReload() {
	app.saveConfig()
	app.loadConfig()
	app.loadServerMap(false)
	app.show()
}

func (app *App) show() {
	Clear()

	// 输出server
	app.showServers()

	// 监听输入
	input, isGlobal := app.checkInput()
	if isGlobal {
		if app.handleGlobalCmd(input) {
			return
		}
	} else {
		server := app.serverIndex[input].server
		Log.Info("select server", server.Name)
		app.TipsMsg(server.Name)
		server.Connect()
	}

}

func (app *App) handleGlobalCmd(cmd string) bool {
	switch strings.ToLower(cmd) {
	case "exit":
		return true
	case "edit":
		app.handleEdit()
		return false
	case "add":
		app.handleAdd()
		return false
	case "remove":
		app.handleRemove()
		return false
	default:
		Errorln("指令无效")
		return false
	}
}

// 编辑
func (app *App) handleEdit() {
	Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		app.show()
		return
	}

	serverIndex, ok := app.serverIndex[id]
	if !ok {
		Errorln("序号不存在")
		app.handleEdit()
		return
	}

	serverIndex.server.Edit()
	app.saveAndReload()
}

// 移除
func (app *App) handleRemove() {
	Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		app.show()
		return
	}

	serverIndex, ok := app.serverIndex[id]
	if !ok {
		Errorln("序号不存在")
		app.handleEdit()
		return
	}

	if serverIndex.indexType == IndexTypeServer {
		servers := app.config.Servers
		app.config.Servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
	} else {
		servers := app.config.Groups[serverIndex.groupIndex].Servers
		servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
		app.config.Groups[serverIndex.groupIndex].Servers = servers
	}

	app.saveAndReload()
}

// 新增
func (app *App) handleAdd() {
	groups := make(map[string]*Group)
	for i := range app.config.Groups {
		group := &app.config.Groups[i]
		groups[group.Prefix] = group
		Info("["+group.Prefix+"]"+group.GroupName, "\t")
	}
	Infoln("[其他值]默认组")
	Info("请输入要插入的组：")
	g := ""
	fmt.Scanln(&g)

	server := Server{}
	server.Format()
	server.Edit()

	group, ok := groups[g]
	if ok {
		group.Servers = append(group.Servers, server)
	} else {
		app.config.Servers = append(app.config.Servers, server)
	}

	app.saveAndReload()
}

// 保存配置文件
func (app *App) saveConfig() error {
	b, err := json.Marshal(app.config)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	err = app.backConfig()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(app.ConfigPath, out.Bytes(), os.ModePerm)
}

func (app *App) backConfig() error {
	srcFile, err := os.Open(app.ConfigPath)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	path, _ := filepath.Abs(filepath.Dir(app.ConfigPath))
	backupFile := path + "/al-" + time.Now().Format("20060102150405") + ".conf"
	desFile, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer desFile.Close()

	_, err = io.Copy(desFile, srcFile)
	if err != nil {
		return err
	}

	Infoln("配置文件已备份：", backupFile)
	return nil
}

// 检查输入
func (app *App) checkInput() (string, bool) {
	flag := ""
	for {
		fmt.Scanln(&flag)
		Log.Info("input scan:", flag)

		if app.isGlobalInput(flag) {
			return flag, true
		}

		if _, ok := app.serverIndex[flag]; !ok {
			for k, v := range app.serverIndex {
				if v.server.Name == flag {
					return k, false
				}
			}
			Errorln("输入有误，请重新输入")
		} else {
			return flag, false
		}
	}
}

// 判断是否全局输入
func (app *App) isGlobalInput(flag string) bool {
	switch flag {
	case "edit":
		fallthrough
	case "add":
		fallthrough
	case "remove":
		fallthrough
	case "exit":
		return true

	default:
		return false
	}
}

// 加载配置文件
func (app *App) loadConfig() {
	b, _ := ioutil.ReadFile(app.ConfigPath)
	err := json.Unmarshal(b, &app.config)
	if err != nil {
		Errorln("加载配置文件失败", err)
		panic(errors.New("加载配置文件失败：" + err.Error()))
	}
}

// 打印列表
func (app *App) showServers() {
	maxlen := app.separatorLength()
	app.formatSeparator(" 欢迎使用 Auto Login ", "=", maxlen)
	for i, server := range app.config.Servers {
		if i%2 == 0 {
			LoglnB(app.recordServer(strconv.Itoa(i+1), server))
		} else {
			LoglnW(app.recordServer(strconv.Itoa(i+1), server))
		}
	}

	for _, group := range app.config.Groups {
		if len(group.Servers) == 0 {
			continue
		}

		app.formatSeparator(" "+group.GroupName+" ", "_", maxlen)
		for i, server := range group.Servers {
			Logln(app.recordServer(group.Prefix+strconv.Itoa(i+1), server))
		}
	}

	app.formatSeparator("", "=", maxlen)
	Logln("", "[add]  添加", "    ", "[edit] 编辑", "    ", "[remove] 删除")
	Logln("", "[exit]\t退出")
	app.formatSeparator("", "=", maxlen)
	Info("请输入序号或操作: ")
}

func (app *App) formatSeparator(title string, c string, maxlength int) {

	charslen := int((maxlength - ZhLen(title)) / 2.0)
	chars := ""
	for i := 0; i < charslen; i++ {
		chars += c
	}

	Infoln(chars + title + chars)
}

func (app *App) separatorLength() int {
	maxlength := 60
	for _, group := range app.config.Groups {
		length := ZhLen(group.GroupName)
		if length > maxlength {
			maxlength = length + 10
		}
	}
	return maxlength
}

// 加载
func (app *App) loadServerMap(check bool) {
	Log.Info("server count", len(app.config.Servers), "group count", len(app.config.Groups))

	for i := range app.config.Servers {
		server := &app.config.Servers[i]
		server.Format()
		flag := strconv.Itoa(i + 1)

		if _, ok := app.serverIndex[flag]; ok && check {
			panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
		}

		server.MergeOptions(app.config.Options, false)
		app.serverIndex[flag] = ServerIndex{
			indexType:   IndexTypeServer,
			groupIndex:  -1,
			serverIndex: i,
			server:      server,
		}
	}

	for i := range app.config.Groups {
		group := &app.config.Groups[i]
		for j := range group.Servers {
			server := &group.Servers[j]
			server.Format()
			flag := group.Prefix + strconv.Itoa(j+1)

			if _, ok := app.serverIndex[flag]; ok && check {
				panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
			}

			server.MergeOptions(app.config.Options, false)
			app.serverIndex[flag] = ServerIndex{
				indexType:   IndexTypeGroup,
				groupIndex:  i,
				serverIndex: j,
				server:      server,
			}
		}
	}
}

func (app *App) recordServer(flag string, server Server) string {
	name := server.Name
	flagMsg := SP[0:1] + "[" + SP[0:3-len(flag)] + flag + "]" + SP[0:3]
	if app.config.ShowDetail {
		name = name + SP[:len(SP)-ZhLen(name)]
		return flagMsg + name + " [" + server.User + "@" + server.IP + ":" + fmt.Sprintf("%d", server.Port) + "]"
	} else {
		return flagMsg + name
	}
}
