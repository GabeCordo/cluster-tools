package provisioner

import "github.com/GabeCordo/cluster-tools/internal/processor/provisioner"

var instance *provisioner.Provisioner

func GetProvisionerInstance() *provisioner.Provisioner {

	if instance == nil {
		instance = provisioner.NewProvisioner()
	}
	return instance
}
