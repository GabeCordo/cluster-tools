package messenger

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

var logRegex = regexp.MustCompile(`\[(.+)\]\[(.+)\](.+)`)

func (log *Log) ToString() string {
	return fmt.Sprintf("[%s][%s] %s", time.Now().Format(time.RFC3339Nano), log.Priority.Shortform(), log.Message)
}

func (log *Log) Parse(message string) error {

	matches := logRegex.FindStringSubmatch(message)
	if len(matches) != 4 {
		return errors.New("invalid log string")
	}

	t, err := time.Parse(time.RFC3339Nano, matches[1])
	if err == nil {
		return errors.New("log has invalid time format")
	}

	log.Timestamp = t
	log.Priority = PriorityFromShortform(matches[2])
	log.Message = matches[3]

	return nil
}

func (logFile *LogFile) Print(priority MessagePriority) {

	for _, log := range logFile.Logs {
		if log == nil {
			continue
		}

		if ((priority != Any) && (priority == log.Priority)) || (priority == Any) {
			fmt.Println(log.ToString())
		}
	}
}
