package processor

import "errors"

var AlreadyExists = errors.New("processor with the host:port already exists")

var DoesNotExist = errors.New("processor with the host:port does not exist")

var ModuleAlreadyRegistered = errors.New("module is already registered to the processor")

var ModuleVersionClash = errors.New("module with same name but different version already exist on the core")

var ModuleContactClash = errors.New("module with same name has different contact information")

var ModuleNotMounted = errors.New("module is not mounted")

var ModuleDoesNotExist = errors.New("module does not exist")

var ClusterNotMounted = errors.New("cluster is not mounted")

var CanNotProvisionStreamCluster = errors.New("stream clusters cannot be called manually like batch processes")

var ClusterDoesNotExist = errors.New("cluster does not exist in the module")
