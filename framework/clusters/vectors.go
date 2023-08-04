package clusters

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"time"

	"github.com/GabeCordo/etl-light/components/channel"
)

type Vector struct {
	x int
	y int
}

func (v Vector) ExtractFunc(m cluster.M, c channel.OneWay) {

	vec := Vector{1, 5} // simulate pulling data from a source
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		c.Push(vec) // send data to the TransformFunc
	}
}

func (v Vector) TransformFunc(m cluster.M, in any) (out any, success bool) {

	vec := (in).(Vector)

	vec.x *= 5
	vec.y += 5

	return vec, true
}

func (v Vector) LoadFunc(m cluster.M, in any) {

	vec := (in).(Vector)
	fmt.Printf("Vec(x: %d, y: %d)\n", vec.x, vec.y)
}

// ---

type VectorWait struct {
	x int
	y int
}

func (v VectorWait) ExtractFunc(m cluster.M, c channel.OneWay) {

	vec := VectorWait{1, 5} // simulate pulling data from a source
	for i := 0; i < 100; i++ {
		c.Push(vec) // send data to the TransformFunc
	}
}

func (v VectorWait) TransformFunc(m cluster.M, in any) (out any, success bool) {

	vec := (in).(VectorWait)

	vec.x *= 5
	vec.y += 5

	return vec, true
}

func (v VectorWait) LoadFunc(m cluster.M, in []any) {

	for _, vec := range in {
		v := (vec).(VectorWait)
		fmt.Printf("Vec(x: %d, y: %d)\n", v.x, v.y)
	}
}
