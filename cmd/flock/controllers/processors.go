package controllers

import (
	"fmt"
	"time"

	"github.com/FortifiedCode/commandline"
	"github.com/FortifiedCode/flock"
	"github.com/FortifiedCode/flock/cmd/flock/local"
	"github.com/FortifiedCode/flock/internal/shared/api"
	"github.com/FortifiedCode/plover"
)

func generator(out chan int) {

	for i := 0; i < 10; i++ {
		out <- 1
	}

	close(out)
}

func add2(a int) (b int) {
	b = a + 2
	time.Sleep(4 * time.Millisecond)
	return b
}

func mul2(a int) (b int) {
	b = a * 2
	time.Sleep(10 * time.Millisecond)
	return b
}

func prt(a int) {
	time.Sleep(1 * time.Millisecond)
	fmt.Println(a)
}

type ProcessorController struct {
}

func (controller ProcessorController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag()
	} else if cli.Flag(commandline.Create) {
		controller.createFlag()
	} else {
		fmt.Println("[!] processors expects a 'show' flag")
	}

	return commandline.Terminate
}

func (controller ProcessorController) showFlag() {

	coreHost := local.GetCore()

	processors, err := api.GetProcessors(coreHost)
	if err != nil {
		fmt.Printf("[!] could not retrieve processors: %v\n", err)
		return
	}

	for _, processor := range processors {
		processor.PrettyPrint()
	}
}

func (controller ProcessorController) createFlag() {

	repository := plover.NewRepository()

	m := repository.Module("common")
	m.Version = "v1.0"
	err := m.LinkFunction("generator", generator)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("add2", add2)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("mul2", mul2)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("prt", prt)
	if err != nil {
		fmt.Print(err)
	}

	processor := flock.New(repository)
	processor.Connect()
}
