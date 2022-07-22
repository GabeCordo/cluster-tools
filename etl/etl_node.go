package etl

import "ETLFramework/net"

func ContainersFunction(request *net.Request, response *net.Response) {

}

func StatisticsFunction(request *net.Request, response *net.Response) {

}

func DebugFunction(request *net.Request, response *net.Response) {

}

func Register() {
	node := GetNodeInstance()
	node.Function("/containers", ContainersFunction, []string{"GET"}, false)
	node.Function("/statistics", StatisticsFunction, []string{"GET"}, true)
	node.Function("/debug", DebugFunction, []string{"GET", "POST", "DELETE"}, true)
}
