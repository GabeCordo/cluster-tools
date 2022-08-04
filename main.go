package main

import (
	"ETLFramework/channel"
	"ETLFramework/core"
	"fmt"
	"log"
	"os"
)

type Vector struct {
	x int
	y int
}

type Multiply struct {
}

func (m Multiply) ExtractFunc(output channel.OutputChannel) {
	v := Vector{1, 5} // simulate pulling data from a source
	output <- v       // send data to the TransformFunc
}

func (m Multiply) TransformFunc(input channel.InputChannel, output channel.OutputChannel) {
	if v, ok := (<-input).(Vector); ok {
		v.x = v.x * 5
		v.y = v.y * 5

		output <- v // send data to the LoadFunc
	}
}

func (m Multiply) LoadFunc(input channel.InputChannel) {
	if v, ok := (<-input).(Vector); ok {
		log.Printf("Vector(%d, %d)", v.x, v.y)
	}
}

func main() {
	// cli flags
	var debug = false

	// start reading cli arguments
	args := os.Args[1:] // strip out the file descriptor in position 0

	for i := range args {
		if args[i] == "-h" {
			fmt.Println("ETLFramework")
			fmt.Println("-h\tView helpful information about the etl service")
			fmt.Println("-d\tEnable debug mode")
			return
		} else if args[i] == "-d" {
			debug = true
		}
	}
	// stop reading cli arguments

	if debug {
		fmt.Println("starting up etlframework..")
	}

	core := core.NewCore()

	m := Multiply{}
	core.Cluster("multiply", m)
	
	core.Run()

	if debug {
		fmt.Println("shutting down etlframework")
	}
}
