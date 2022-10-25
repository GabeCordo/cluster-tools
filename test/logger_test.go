package main

import (
	"etl/logger"
	"testing"
	"time"
)

func TestLoggerAlert(t *testing.T) {
	l := logger.NewLogger(".", logger.Verbose, logger.NewInterval(0, 1))

	go l.LoggerEventLoop()

	l.Alert("Hi")

	time.Sleep(1 * time.Second)
}
