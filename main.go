package main

import (
	"ETLFramework/channel"
	"ETLFramework/cli"
	"ETLFramework/cluster"
	"ETLFramework/core"
	"log"
	"time"
)

type Vector struct {
	x int
	y int
}

type Multiply struct {
}

func (m Multiply) ExtractFunc(output channel.OutputChannel) {
	v := Vector{1, 5} // simulate pulling data from a source
	for i := 0; i < 10; i++ {
		output <- v // send data to the TransformFunc
	}
	close(output)
}

func (m Multiply) TransformFunc(input channel.InputChannel, output channel.OutputChannel) {
	for request := range input {
		if v, ok := (request).(Vector); ok {
			v.x = v.x * 5
			v.y = v.y * 5

			output <- v // send data to the LoadFunc
		}
		time.Sleep(500 * time.Millisecond)
	}
	close(output)
}

func (m Multiply) LoadFunc(input channel.InputChannel) {
	for request := range input {
		if v, ok := (request).(Vector); ok {
			log.Printf("Vector(%d, %d)", v.x, v.y)
		}
	}
}

func main() {
	c := core.NewCore()

	m := Multiply{}
	c.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

	if commandLine, ok := cli.NewCommandLine(c); ok {
		commandLine.Run()
	}
}
