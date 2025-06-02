package local

func GetNamespace() string {

	config := Config{}
	if err := getOrCreateConfig(&config); err != nil {
		panic(err)
	}

	return config.Namespace
}

func SwitchNamespace(namespace string) {

	config := Config{}
	if err := getOrCreateConfig(&config); err != nil {
		panic(err)
	}

	config.Namespace = namespace

	if err := updateConfig(&config); err != nil {
		panic(err)
	}
}

func GetCore() string {

	config := Config{}
	if err := getOrCreateConfig(&config); err != nil {
		panic(err)
	}

	return config.Core
}

func SwitchCore(core string) {

	config := Config{}
	if err := getOrCreateConfig(&config); err != nil {
		panic(err)
	}

	config.Core = core

	if err := updateConfig(&config); err != nil {
		panic(err)
	}
}
