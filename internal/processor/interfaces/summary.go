package interfaces

type SupervisorSummary struct {
	Module     string
	Cluster    string
	Supervisor uint64
	Statistics *Statistics
	ETState    string
	ETSize     int
	TLState    string
	TLSize     int
}
