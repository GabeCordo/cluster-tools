package core

import (
	"fmt"
	"time"
)

type LoggerOutput int

const (
	Verbose LoggerOutput = iota
	Simple
)

type ILogger interface {
	Log(template string, params ...interface{})
	Alert(template string, params ...interface{})
	Warning(template string, params ...interface{})
}

type NodeLogger struct {
	folder string
	style  LoggerOutput
	node   *Node
}

func NewLogger(folder string, style LoggerOutput, node *Node) NodeLogger {
	return NodeLogger{folder, style, node}
}

func (l NodeLogger) Log(template string, params ...interface{}) {
	dt := time.Now().Format("01-02-2006 15:04:05")
	header := fmt.Sprintf("[%s][%s]", dt, l.node.name)
	data := fmt.Sprintf(template, params...)
	fmt.Println(header + " " + data)
}

func (l NodeLogger) Alert(template string, params ...interface{}) {
	data := fmt.Sprintf("[!] %s", params...)
	fmt.Println(data)
}

func (l NodeLogger) Warning(template string, params ...interface{}) {
	data := fmt.Sprintf("[?] %s", params...)
	fmt.Println(data)
}
