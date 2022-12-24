package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// INTERACTIVE DASHBOARD START

type InteractiveDashboardCommand struct {
	PublicName string
}

func (idc InteractiveDashboardCommand) Name() string {
	return idc.PublicName
}

func (idc InteractiveDashboardCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block until we receive an interrupt from the system
		fmt.Println()
		os.Exit(0)
	}()

	for {
		now := time.Now()
		fmt.Printf("%d:%d:%d\r", now.Hour(), now.Minute(), now.Second())

		time.Sleep(1 * time.Second)
	}

	return true
}
