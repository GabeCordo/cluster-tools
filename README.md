# ETLFramework

### CLI Parameters

#### -h --help
Provides descriptions for available parameters.

#### -d --debug
Provides verbose output for the starter script.

#### -g --generate-key
ECDSA public and private keys are outputted as x509 encoded formats.

---

### Common Questions

#### What is the ETLFramework Core?
The core is defined as the entrypoint to the ETLFramework that allows developers to register new Clusters and inject custom configurations
using the *config.etl.json file*.

```go
c := core.NewCore()

m := Multiply{} 	// A structure implementing the ETLFramework.Cluster.Cluster interface
c.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

c.Run()	 // Starts the ETLFramework
```

#### What is a Cluster?
A cluster is defined as any structure that implements the ETLFramework.cluster.Cluster interface. Where the interface can be thought as 
the set of functions required to implement the business-logic of the Extract-Transform-Load (ETL) process. 

```go
type Cluster interface {
    ExtractFunc(output channel.OutputChannel)
    TransformFunc(input channel.InputChannel, output channel.OutputChannel)
    LoadFunc(input channel.InputChannel)
}
```

#### What does the ETLFramework do with a Cluster?
Once a cluster has been registered with the ETLFramework Core, it can be mounted and provisioned to initiate execution. Where an ETLCluster is linked by
go channels to pass data between the successive functions. The framework is responsible for monitoring the amount of data present within the channels, and if required, provisioning
additional thread to assist with data-processing.

```go
extract -[etChannel]-> transform -[tlChannel]-> load
```

etChannel : the channel between the (extract) and (transform) goroutines
tlChannel : the channel between the (transform) and (load) goroutines

##### How is provisioning handled?

Each channel (et and tl) has an associated threshold and growth factor. The developer has the option of specifying these quanities to
best match the ETL-process they are implementing or used the default as defined by the ETLFramework.

- Data Unit: A single object or structure past to the channel that is required by the next E-T-L function.
- Threshold (int): When 
- Growth Factor (double): By what scale should the number of successive functions exist if a channel is considered "congested"

#### Is Cluster Execution Guaranteed?

Each channel's completion is guaranteed by synchronous Wait Groups, where the ETLFramework Core will not
complete until every Cluster has completed processing.

#### How is Cluster Deadlock Prevented?

In theory, each channel has an infinite runtime until an operator has made a request to shut down the ETLFramework Core. Upon receiving the
shutdown interrupt, each Cluster will have 30 minutes to finish executing before being terminated. If this value does not fit your defined
scope, it can be modified in the *config.etl.json* under the "hard-terminate-time" flag as an integer representation of minutes.

### Provisioning a Cluster

#### What is Mounting?

Mounting indicates whether a cluster can be dynamically provisioned. When a cluster is registered
with the ETLFramework Core it is placed in a "registered" state, but is not operational. In order to
become operational where the cluster can be provisioned, it must be "mounted" to be placed into an operational state.

#### Why is Mounting important?

Mounting allows the operator to control what clusters are available during the lifetime of the system. When a cluster
is using excessive resources, is encountering unexpected errors, or has a possible vulnerability the operator can dismount
the cluster to stop further provisioning.

---

### HTTP Curl Interaction

#### Endpoints

##### /clusters
1. provision
2. mount
3. unmount

##### /data

1. mounts
2. supervisor
   1. lookup
   2. state

##### /statistics
1. [cluster-name]

##### /debug
1. shutdown
2. endpoints
   1. [endpoint-identifier]

#### Examples

##### Shutdown Node
curl -X GET http://127.0.0.1:8000/debug -H 'Content-Type: application/json' -d '{"function": "shutdown"}'

##### Test Cluster

###### View Mounted and Unmounted Clusters
curl -X GET http://127.0.0.1:8000/data -H 'Content-Type: application/json' -d '{"function": "mounts"}'

###### Mount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "mount", "param":["multiply"]}'

###### UnMount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "unmount", "param":["multiply"]}'

###### Provision Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "provision", "param":["multiply"]}'

##### Test Cluster Statistics
curl -X GET http://127.0.0.1:8000/statistics -H 'Content-Type: application/json' -d '{"function": "multiply", "param":["multiply"]}'