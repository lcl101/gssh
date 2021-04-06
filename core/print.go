package core

import (
	"fmt"
	"runtime"
)

// 打印一行信息
// 字体颜色为默色
func Logln(a ...interface{}) {
	fmt.Println(a...)
}

// 打印一行信息
// 背景色为黑
func LoglnB(a ...interface{}) {
	if isNotWin() {
		fmt.Print("\033[36m")
	}
	fmt.Println(a...)
	if isNotWin() {
		fmt.Print("\033[0m")
	}
}

// 打印一行信息
// 背景色为白
func LoglnW(a ...interface{}) {
	// if isNotWin() {
	// 	fmt.Print("\033[30;47m")
	// }
	fmt.Println(a...)
	// if isNotWin() {
	// 	fmt.Print("\033[0;0m")
	// }
}

// 打印一行信息
// 字体颜色为绿色
func Infoln(a ...interface{}) {
	if isNotWin() {
		fmt.Print("\033[32m")
	}
	fmt.Println(a...)
	if isNotWin() {
		fmt.Print("\033[0m")
	}
}

// 打印信息（不换行）
// 字体颜色为绿色
func Info(a ...interface{}) {
	if isNotWin() {
		fmt.Print("\033[32m")
	}
	fmt.Print(a...)
	if isNotWin() {
		fmt.Print("\033[0m")
	}
}

// 打印一行错误
// 字体颜色为红色
func Errorln(a ...interface{}) {
	if isNotWin() {
		fmt.Print("\033[31m")
	}
	fmt.Println(a...)
	if isNotWin() {
		fmt.Print("\033[0m")
	}
}

func isNotWin() bool {
	return runtime.GOOS != "windows"
}
