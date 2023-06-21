package main

import (
	"fmt"

	"github.com/GabeCordo/etl-light/components/channel"
)

type Vector struct {
	x int
	y int
}

func (m Vector) ExtractFunc(c channel.OneWay) {

	v := Vector{1, 5} // simulate pulling data from a source
	for i := 0; i < 100; i++ {
		c.Push(v) // send data to the TransformFunc
	}
}

func (m Vector) TransformFunc(in any) (out any, success bool) {

	v := (in).(Vector)

	v.x *= 5
	v.y += 5

	return v, true
}

func (m Vector) LoadFunc(in any) {

	v := (in).(Vector)
	fmt.Printf("Vec(x: %d, y: %d)\n", v.x, v.y)
}
