package clusters

import (
	"fmt"
	"github.com/GabeCordo/mango/components/channel"
	"github.com/GabeCordo/mango/components/cluster"
	"github.com/GabeCordo/mango/utils"
)

type MetaDataCluster struct {
	helper utils.Helper
}

func (mdc *MetaDataCluster) SetHelper(helper utils.Helper) {
	mdc.helper = helper
}

func (mdc *MetaDataCluster) ExtractFunc(m cluster.M, c channel.OneWay) {

	key := m.GetKey("test")

	if key == "" {
		fmt.Println("key was not passed successfully")
	} else {
		c.Push(key)
	}
}

func (mdc *MetaDataCluster) TransformFunc(m cluster.M, in any) (out any, success bool) {
	return in, true
}

func (mdc *MetaDataCluster) LoadFunc(m cluster.M, in any) {
	key := (in).(string)
	mdc.helper.Log(key)
}
