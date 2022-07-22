package main

import (
	"ETLFramework/channel"
	"ETLFramework/etl"
	"log"
	"testing"
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

func TestRegisterETLGroup(t *testing.T) {
	multiply := Multiply{}

	monitor := etl.NewMonitor(multiply)
	monitor.Start()
}
