package blossom_example

import cluster_tools "github.com/GabeCordo/cluster-tools"

func add(a, b int) (c int) {
	c = a + b
	return c
}

func mulByTwo(a int) (b int) {
	b = a * 2
	return b
}

func main() {

	cfg := cluster_tools.NewConfig("tmp")
	prscr, _ := cluster_tools.New(cfg)

	commonModule := prscr.Module("common")
	commonModule.Mounted = true

	addFunction, _ := commonModule.Function("add", add)
	addFunction.Mounted = true

	mulFunction, _ := commonModule.Function("mul", mulByTwo)
	mulFunction.Mounted = true

	prscr.Run()
}
