package provisioner

import "github.com/GabeCordo/clarence/internal/components/provisioner"

var instance *provisioner.Provisioner

func GetProvisionerInstance() *provisioner.Provisioner {

	if instance == nil {
		instance = provisioner.NewProvisioner()
	}
	return instance
}
