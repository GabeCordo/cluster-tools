package main

import (
	"fmt"
	cluster_tools "github.com/GabeCordo/cluster-tools"
)

func generator(out chan int) {

	for i := 0; i < 10; i++ {
		out <- i
	}
}

func add2(a int) (b int) {
	b = a + 2
	return b
}

func mul2(a int) (b int) {
	b = a * 2
	return b
}

func prt(a int) {
	fmt.Println(a)
}

func main() {

	cfg := cluster_tools.NewConfig("tmp")
	p, _ := cluster_tools.New(cfg)

	m := p.Module("common")
	m.LinkFunction("generator", generator)
	m.LinkFunction("add2", add2)
	m.LinkFunction("mul2", mul2)
	m.LinkFunction("prt", prt)

	p.Connect("http://localhost:8137")
	defer p.Disconnect()

	p.Run()
}
