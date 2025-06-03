package controllers

import (
	"fmt"
	"time"

	cluster_tools "github.com/GabeCordo/Flock"
	"github.com/GabeCordo/commandline"
)

func generator(out chan int) {

	for i := 0; i < 1000000; i++ {
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

type ExampleController struct {
}

func (controller ExampleController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	p, _ := cluster_tools.New()

	m := p.Module("common")
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

	p.Runtime()

	return commandline.Terminate
}
