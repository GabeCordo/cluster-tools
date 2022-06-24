package logger

import (
	"fmt"
	"os"
	"time"
)

func NewLogger(folder string, style LogOutput, interval *LogInterval) *Logger {
	// check to see if the path provided by the logger exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil
	}
	channel := make(chan string)
	logger := new(Logger)

	logger.Folder = folder
	logger.Style = style
	logger.LogQueue = channel
	logger.Interval = interval

	return logger
}

func (l *Logger) LoggerEventLoop() {
	for {
		// block until we get a new log request sent from the log/warning/alert functions
		log := <-l.LogQueue

		// we need to check if the path exists in-case the struct was
		// created without the use of a constructor
		if _, err := os.Stat(l.Folder); os.IsNotExist(err) {
			// discard the log
			os.Exit(-1) // TODO - replace this
		}

		if l.Interval.Expired() {
			l.Interval.Refresh()
		}

		fileName := l.Interval.String()
		file, err := os.OpenFile(l.Folder+fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			os.Exit(-1) // TODO - replace this
		}

		if l.Style == Verbose {
			fmt.Println(log)
		}
		if _, err := file.Write([]byte(log + "\n")); err != nil {
			os.Exit(-1) // TODO - replace this
		}
		if err := file.Close(); err != nil {
			os.Exit(-1) // TODO - replace this
		}
	}
}

func (l *Logger) Init() {
	go l.LoggerEventLoop() // a goroutine is created to print/store logs independent of
	// the main execution thread
}

func (l *Logger) Log(objectName, template string, params ...interface{}) {
	var data, header string
	if l.Style == Verbose {
		dt := time.Now().Format(dateOutputFormat)
		header = fmt.Sprintf("[%s][%s]", dt, objectName)
		data = fmt.Sprintf(template, params...)
	} else {
		data = fmt.Sprintf(template, params...)
	}
	l.LogQueue <- header + " " + data
}

func (l *Logger) Alert(template string, params ...interface{}) {
	data := fmt.Sprintf("[!] %s", params...)
	l.LogQueue <- data
}

func (l *Logger) Warning(template string, params ...interface{}) {
	data := fmt.Sprintf("[?] %s", params...)
	l.LogQueue <- data
}
