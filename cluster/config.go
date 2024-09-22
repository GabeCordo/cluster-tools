package cluster

type EtlMode string

const (
	Batch  EtlMode = "mode/batch"  // The cluster is provisioned when invoked by an operator or application.
	Stream         = "mode/stream" // The cluster is provisioned automatically when the system is started.
)

type OnCrash string

const (
	Restart   OnCrash = "Restart"
	DoNothing         = "DoNothing"
)

type OnLoad string

const (
	CompleteAndPush OnLoad = "CompleteAndPush"
	WaitAndPush            = "WaitAndPush"
)

type Status uint8

const (
	Registered = iota
	UnMounted
	Mounted
	InUse
	MarkedForDeletion
)

// M contains metadata about the running supervisor including any state
// information that a developer might need to interact with the Supervisor.
type M interface {
	GetKey(key string) string
}

type H interface {
	IsDebugEnabled() bool
	SaveToCache(data string) (string, error)
	LoadFromCache(identifier string) (string, error)
	Log(message string) error
	Logf(format string, data ...any) error
	Warning(message string) error
	Warningf(format string, data ...any) error
	Fatal(message string) error
	Fatalf(format string, data ...any) error
}

type Out interface {
	Push(any) bool
}

// Cluster is a set of functions that define an ETL process
// cluster functions are provisioned on goroutines to run in parallel and
// process data.
type Cluster interface {
	ExtractFunc(helper H, metadata M, out Out)
	TransformFunc(helper H, metadata M, in any) (out any, success bool)
}
