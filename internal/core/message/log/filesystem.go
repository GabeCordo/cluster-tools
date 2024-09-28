package log

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/message"
	"strings"
)

type File struct {
	Logs      []*Log
	NumOfLogs int
}

func (logFile *File) Print(priority message.Priority) {

	for _, l := range logFile.Logs {
		if l == nil {
			continue
		}

		if ((priority != message.Any) && (priority == l.Priority)) || (priority == message.Any) {
			fmt.Println(l.ToString())
		}
	}
}

func NewLogFile(bytes []byte) *File {
	instance := new(File)

	logs := strings.Split(string(bytes), "\n")
	instance.NumOfLogs = len(logs) - 1 // there will always be an empty split due to the final \n
	instance.Logs = make([]*Log, instance.NumOfLogs)

	for idx, logStr := range logs {
		lg := NewLog()
		if err := lg.Parse(logStr); err == nil {
			instance.Logs[idx] = lg
		}
	}

	return instance
}
