package scp

import (
	"fmt"
	"io"

	"github.com/redmask-hb/GoSimplePrint/goPrint"
)

var (
	msg = make(chan *View, 10)
)

func init() {
	//启动一个显示线程
	go viewMsg()
}

type View struct {
	name    string
	percent int
}

func NewView(name string, per int) *View {
	return &View{
		name:    name,
		percent: per,
	}
}

func (v *View) SetPercent(pre int) {
	v.percent = pre
}

func viewMsg() {
	bar := goPrint.NewBar(100)
	bar.SetGraph(">")
	bar.HideRatio()
	for {
		v := <-msg
		if v.name == "EOF" {
			fmt.Println("")
		}
		// fmt.Println("xxxxxxx")
		//显示
		bar.SetNotice(v.name)
		bar.PrintBar(v.percent)
		if v.percent >= 100 {
			fmt.Println("")
		}
	}
}

type ProxyReader struct {
	r    io.Reader
	size int
	comp int
	name string
}

func NewProxyReader(r io.Reader, name string, size int) *ProxyReader {
	return &ProxyReader{
		r:    r,
		size: size,
		comp: 0,
		name: name,
	}
}

func (proxy *ProxyReader) Read(p []byte) (n int, err error) {
	n, err = proxy.r.Read(p)
	if err != nil {
		msg <- NewView("EOF", proxy.comp/proxy.size)
		return
	}
	proxy.comp += n
	msg <- NewView(proxy.name, proxy.comp*100/proxy.size)
	return
}

type ProxyWriter struct {
	w    io.Writer
	size int
	comp int
	name string
}

func NewProxyWriter(w io.Writer, name string, size int) *ProxyWriter {
	return &ProxyWriter{
		w:    w,
		size: size,
		comp: 0,
		name: name,
	}
}

func (proxy *ProxyWriter) Write(p []byte) (n int, err error) {
	n, err = proxy.w.Write(p)
	if err != nil {
		msg <- NewView("EOF", proxy.comp/proxy.size)
		return
	}
	proxy.comp += n
	msg <- NewView(proxy.name, proxy.comp*100/proxy.size)
	return
}
