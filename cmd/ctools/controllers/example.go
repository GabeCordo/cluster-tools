package controllers

import (
	"fmt"
	cluster_tools "github.com/GabeCordo/cluster-tools"
	"github.com/GabeCordo/commandline"
	"time"
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

	cfg := cluster_tools.NewConfig("tmp")
	p, _ := cluster_tools.New(cfg)

	m := p.Module("common")
	m.LinkFunction("generator", generator)
	m.LinkFunction("add2", add2)
	m.LinkFunction("mul2", mul2)
	m.LinkFunction("prt", prt)

	p.Run()

	return commandline.Terminate
}
