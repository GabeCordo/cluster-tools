package clusters

import (
	"fmt"
	"github.com/GabeCordo/mango/components/channel"
	"github.com/GabeCordo/mango/components/cluster"
	"github.com/GabeCordo/mango/utils"
	"time"
)

type Vector2D struct {
	x int
	y int
}

// --

type VectorCluster struct {
	helper utils.Helper
}

func (v *VectorCluster) SetHelper(helper utils.Helper) {
	v.helper = helper
}

func (v *VectorCluster) ExtractFunc(m cluster.M, c channel.OneWay) {

	vec := Vector2D{1, 5} // simulate pulling data from a source
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Millisecond)
		c.Push(vec) // send data to the TransformFunc
	}
}

func (v *VectorCluster) TransformFunc(m cluster.M, in any) (out any, success bool) {

	vec := (in).(Vector2D)

	vec.x *= 5
	vec.y += 5

	return vec, true
}

func (v *VectorCluster) LoadFunc(m cluster.M, in any) {

	vec := (in).(Vector2D)
	output := fmt.Sprintf("Vec(x: %d, y: %d)\n", vec.x, vec.y)
	v.helper.Log(output)
}

// ---

type VectorWaitCluster struct {
	helper utils.Helper
}

func (v *VectorWaitCluster) SetHelper(helper utils.Helper) {
	v.helper = helper
}

func (v *VectorWaitCluster) ExtractFunc(m cluster.M, c channel.OneWay) {

	vec := Vector2D{1, 5} // simulate pulling data from a source
	for i := 0; i < 100; i++ {
		c.Push(vec) // send data to the TransformFunc
	}
}

func (v *VectorWaitCluster) TransformFunc(m cluster.M, in any) (out any, success bool) {

	vec := (in).(Vector2D)

	vec.x *= 5
	vec.y += 5

	return vec, true
}

func (v *VectorWaitCluster) LoadFunc(m cluster.M, in []any) {

	for _, vec := range in {
		v2d := (vec).(Vector2D)
		output := fmt.Sprintf("Vec(x: %d, y: %d)\n", v2d.x, v2d.y)
		v.helper.Log(output)
	}
}
