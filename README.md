# ETLFramework
A software orchestration framework for extract-transform-load process deployment and scaling. Developers can write and link custom ETL functions for data processing, that will be provisioned and scaled according to data velocity and processing demand made by the deployed functions. In production, ETL functions can be provisioned manually (or by script) through function calls over RPC using the "fast backend" framework. ETL processes can be mounted or unmounted depending on whether the administrator wishes to allow RPC calls to provision new instances of the ETL process.

### Features

1. Config for persistent and configurable data
2. Auth endpoints can be dynamically added to the system upon startup through the config
3. Dynamic provisioning of cluster functions
4. Statistics and Usage data storage for clusters called during the session
5. Clean Teardown
   - even if the system has a SIGINT called or has been requested to shutdown, the system will not complete until every started ETL process completes
6. Deadlock avoidance with ETL clusters
   - as it stands, if a cluster is in deadlock and the system must wait for it to complete before terminating, the system will never shutdown
7. Guarantee that all queues are cleaned up + processed before completion
8. Dynamic provisioning of Clusters over RPC
9. Static Mounting of Clusters
   - before a Cluster can be used, it must be marked as "mounted" in the config files. This is to avoid newly added functions (that may be under review) being automatically deployed into production.
11. Mounting and Demounting Clusters for Use During Runtime
   - in the case that a Cluster is triggering errors or runtime issues, it may be dismounted such that it cannot be invoked further.

### Common Questions

#### What is the ETLFramework Core?
The core is defined as the entrypoint to the ETLFramework that allows developers to register new Clusters and inject custom configurations
using the *config.etl.json file*.

```go
c := core.NewCore()

m := Multiply{} 	// A structure implementing the etl.Cluster.Cluster interface
c.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

c.Run()	 // Starts the etl
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

### ETLHelper
The ETLHelper is a structure containing useful functions for interacting with components of the ETL Framework dynamically.


The ETLHelper must be passed a pointer to an ETLCore on initialization. This means that an ETLHelper can also be used to interact
with more than one ETLCore.
```go
    etl := core.NewCore()
    helper := core.NewHelper(etl)
```


#### Cache
The ETL Framework has a built-in solution for caching intermediate data that is used in a chain of sequential ETL clusters.
This is an alternative to loading and storing data in an external database solution that introduces networking errors and latency.

##### Ensuring the Cache is Enabled
When starting up the ETL Framework on debug mode, you should see a system log notifying you that the cache thread has started.

```2023/01/25 12:37:45 (+) Cache Thread Started```

##### Cache Lifetime
Cache data is intended for short lifetime storage. This is to handle that the cache resides in system memory (which is usually limited) and
that permanent or long-term storage of data should be left to other database solutions.

**By Default Cache Records Have a Lifetime of 1 Minute**

###### Modifying the Cache Lifetime
The ETLConfig contains am "expire-in" field which specifies how many minutes should pass before a record is automatically removed from the cache.

```json
{
   "cache": {
      "expire-in": 1
   }
}
```

##### Cache Memory Usage
The cache resides in the system memory which means that the storage capabilities are highly limited versus traditional
database solutions that use disk storage. It is important we keep track of the number of cached values we store on the
ETL Cache in case developers don't realize how much data their clusters are generating, or it is being misused as a black
box for data storage.

**By Default Cache Records Have a Limit of 1000 Records**
If we have records storing 1MB of data, that is already using 1GB of system memory if maxed out.

###### Modifying the Memory Usage
The ETLConfig contains a "max-size" field in the "cache" section to change the max number of records allowed.

```json
{
   "cache": {
      "max-size": 1000
   }
}
```

##### Cache Functions
The following are callable functions that provide access to the ETLCache thread.

###### CachePromise
All functions called to the ETLCache return an ETLCachePromise, an event listener connected to the Cache, that can
block execution until a response is received. *It is important to note that messages sent to and from the cache are asynchronous, meaning
that the response can be anywhere from instantaneous to one-hundred microseconds.

- Calling **Wait()** on a CachePromise blocks execution until a response is received from the ETLCache. Wait() will return an
ETLCacheResponse message which can be used to understand what is stored in the cache.

```go
type CacheResponse struct {
	Identifier string
	Nonce      uint32
	Data       any
	Success    bool
}
```

###### SaveToCache(*data* **any**)
Pass data of any type to be stored locally in the cache. The *CacheResponse* received will contain an **Identifier** assigned by the cache to pull data from the *LoadFromCache* function.

###### LoadFromCache(*identifier* **string**)
Pull data from the cache using the **Identifier** received when storing the data to the cache. The *CacheResponse* received will contain a **Data** value from the local storage. 

- If the data has expired, the **Success** field in the *CacheResponse* object will be **False** 

##### Cache Example

```go
   etl := core.NewCore()
    helper := core.NewHelper(etl)
   
   /
   promise := helper.SaveToCache("test data")
   response := promise.Wait()
   
   promise = helper.LoadFromCache(response.Identifier)
   response = promise.Wait()
```

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

###### Get Cluster Running State
curl -X GET http://127.0.0.1:8000/data -H 'Content-Type: application/json' -d '{"function": "supervisor", "param":["state", "vector"]}'

Expected Output
```json
{"status":200,"description":"no error","data":{"1":"Terminated"}}
```

###### Lookup Cluster
curl -X GET http://127.0.0.1:8000/data -H 'Content-Type: application/json' -d '{"function": "supervisor", "param":["lookup", "vector"]}'

Expected Output
```json
{"status":200,"description":"no error"}
```

###### View Mounted and Unmounted Clusters
curl -X GET http://127.0.0.1:8000/data -H 'Content-Type: application/json' -d '{"function": "mounts"}'

Expected Output
```json
{"status":200,"description":"no error","data":{"vector":true}}
```

###### Mount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "mount", "param":["multiply"]}'

###### UnMount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "unmount", "param":["multiply"]}'

###### Provision Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "provision", "param":["multiply"]}'

##### Cluster Statistics
curl -X GET http://127.0.0.1:8000/statistics -H 'Content-Type: application/json' -d '{"function": "first-pass"}'

Expected Output
```json
{"status":200,"data":{"value":[{"timestamp":"2023-01-28T12:43:12.289353-05:00","elapsed":6640546675,"statistics":{"num-provisioned-extract-routines":1,"num-provisioned-transform-routes":1,"num-provisioned-load-routines":1,"num-et-threshold-breaches":0,"num-tl-threshold-breaches":0}}]}}
```
