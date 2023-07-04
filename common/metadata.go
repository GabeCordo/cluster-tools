package common

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/channel"
	"github.com/GabeCordo/etl-light/components/cluster"
)

type MetaDataCluster struct {
}

func (mdc MetaDataCluster) ExtractFunc(m cluster.M, c channel.OneWay) {

	key := m.GetKey("test")

	if key == "" {
		fmt.Println("key was not passed successfully")
	} else {
		c.Push(key)
	}
}

func (mdc MetaDataCluster) TransformFunc(m cluster.M, in any) (out any, success bool) {
	return in, true
}

func (mdc MetaDataCluster) LoadFunc(m cluster.M, in any) {
	key := (in).(string)
	fmt.Println(key)
}
