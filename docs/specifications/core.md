# core
The core is a centralized service for receiving, distributing, and monitoring requests to run data pipelines.

The core was designed to simplify the deployment and monitoring of data pipelines.


---

## Users
The users of the core are the [operator](#operator-user) and [developer](#developer-user). It's important to
note that these two users may be the same person depending on the size of the team.

### Operator User
The operator is responsible for the production deployment and monitoring of data pipelines.

The operator will be abbreviated as "O" within the document.

### Developer User
The developer is responsible for building and iterating on data pipelines.

The developer will be abbreviated as "D" within the document.

### Product Owner User
The product owner is responsible for the completion of features and billing. 

The product owner will be abbreviated as "PO" within the document.

---


## Goals
Goals are prefixed with "G" followed by the abbreviation of the user and unique identifier per user.

| Identifier | Status   | Achieved In Version | Link |
| :- |:---------|:--------------------|:-----|
| G01 | achieved | todo                |  |
| G02 | missing |                  |      |
| G03 | achieved | todo | |
| G04 | missing |  | |
| GD1 | achieved | | | |
| GD2 | achieved | | | |
| GD3 | missing | | | |
| GP1 | missing | | | |
| GP2 | missing | | | |
| GP3 | missing | | | |
_Table. The goals of the core process._

### Operator Goals
The goals of the [operator](#operator-user) user.

###### GO1
Make a pipeline available to external users and applications.

###### GO2
Know when pipelines have an uptick in faults.

###### GO3
Disable pipelines that show an uptick in faults.

###### GO4
Rollback pipelines to a previous working version when the latest has issues.

### Developer Goals
The goals of the [developer](#developer-user) user.

###### GD1
Develop data pipelines that process .

###### GD2
Monitor data flow through the pipeline to find bottlenecks.

###### GD3
View logs generated as data flows through the pipeline.

###### GD4
Have the resources of the data pipeline automatically scale to meet the demand of the data flowing through the pipeline.

### Product Owner Goals
The goals of the [product owner](#product-owner-user) user.

###### GP1
Assign key performance metrics to a data pipeline. 

###### GP2
View key performance metrics to a data pipeline.

###### GP3
Know when key performance metrics for a data pipeline have been breached.

---


## Requirements
Requirements are prefixed with "R" followed by a unique identifier per requirement.

| Identifier | Achieved | Achieved in Version |
| :- |:---------|:--------------------|
| R0 | Th

###### R0
The core shall represent executable code as a function.

###### R1
The core shall store the parameter types of a function.

###### R2
The core shall store the types returned from a function.

###### R3
The core shall represent a set of functions as a module.

###### R4
The core shall version each module to track incremental changes.

###### R5
The core shall define code that can be run as mounted. 

###### R6
The core shall define code that cannot be run as unmounted.

###### R7
The core shall allow an operator to mount a module.

###### R8
The core shall allow an operator to unmount a module.

###### R9
The core shall allow an operator to mount a function.

###### R10
The core shall allow an operator to unmount a function.

###### R11
The core shall define a set of functions in a pipeline.

###### R12
The core shall define a set of pipes in a pipeline.

###### R13
The core may define a pipe a function receives data from in a pipeline.

###### R14
The core may define a pipe that a function sends data to in a pipeline.

###### R15
The core shall define a pipeline as the smallest executable unit.

> When send a request for code to run on the core, we are sending a request to run a pipeline rather than a singular or set of functions. 

###### R16
The core shall verify each function in a pipeline is mounted before code is executed.

###### R17
The core shall find a processor that supports each function in a pipeline so code may be executed.

###### R18
The core shall send a request to a [processor](#processor) to run a [pipeline](#pipeline). 

---

## Models
These models are standardized representations to store data inside the core.

### Processor
A processor is an external computer process that connects to the core to provide compute resources.

The processor exposes [modules](#module) that contain a set of runnable [functions](#function) on the core.

###### Processor Status
```json
[
  "active"
]
```

###### Processor Record
```json
{
  "Id": "string",
  "RemoteAddr": "string",
  "Status": "ProcessorStatus",
  "LastUpdate": "time",
  "Modules": "list[str]",
  "Retries": 0,
  "NumOfRuns": 0
}
```

### Module

### Function

These modules and functions can be [mounted](#mount-function) or [unmounted](#unmount-function) to
control whether they can be run on a processor.


### Pipeline

### Run

### Statistic



---



## HTTP APIs
Defines the endpoints used to query information inside the core.

### /processor

#### GET
Fetch a set of [processor](#11-processor) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/processor
```

---


### /module

#### GET
Fetch a set of [module](#12-module) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/module
```

#### PUT

##### Mount Module
Allow an operator to run all functions inside a module. 

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/module -H "Content-Type: application/json" -d "{\"module\":\"common\",\"mounted\":true}"
```

##### Unmount Module
Disable an operator from running all functions inside a module. 

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/module -H "Content-Type: application/json" -d "{\"module\":\"common\",\"mounted\":false}"
```

---


### /function
Fetch a set of [function](#13-function) records from the core.

#### GET

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/function\?module=common
```

#### PUT

##### Mount Function
Allow an operator to run a function inside a module.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/function -H "Content-Type: application/json" -d "{\"module\":\"common\",\"function\":\"prt\",\"mounted\":true}"
```

##### Unmount Function
Disable an operator from running a function inside a module.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/function -H "Content-Type: application/json" -d "{\"module\":\"common\",\"function\":\"prt\",\"mounted\":false}"
```

---


### /namespaces

#### GET
Fetch a set of namespaces used in the core. Namespaces are used to allow scoped use of pipeline and job identifiers rather
than maintaining a global namespace.

The _common_ namespace is the default global namespace that is present in all cores.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/namespaces
```


---


### /pipeline

#### GET

###### HTTP Params
namespace 

pipeline _(optional)_

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/pipeline\?namespace=common
```

#### POST
Create a new pipeline that defines the flow of data between functions run on processors.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/pipeline -H "Content-Type: application/json" -d @docs/examples/pipelines/hello-world.json
```

#### PUT
Update a pipeline stored on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/pipeline -H "Content-Type: application/json" -d @docs/examples/pipelines/hello-world.json
```

#### DELETE
Delete a pipeline stored on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X DELETE -L http://127.0.0.1:8136/pipeline\?namespace=common\&pipeline=hello-world
```


---


### /run

#### GET
Fetch the data received from a processor for a running pipeline.

###### HTTP Params
namespace (mandatory)

pipeline (optional)

maximumResults (optional)

offsetOfResults (optional)

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run\?namespace=common\&pipeline=hello-world
```

#### POST
Provision a new running instance of a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/run -H "Content-Type: application/json" -d @docs/examples/runs/hello-world.json
```

### /run/count

#### GET
Get the number of runs for a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/count\?namespace=common\&pipeline=hello-world
```

---


### /statistic

#### GET
Fetch the statistics collected from completed pipeline runs.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/statistic\?namespace=common\&pipeline=hello-world
```


---


### /statistic/info

#### GET

##### Retrieve Namespaces for Statistics

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/statistic/info
```

##### Retrieve Pipeline Statistics for a Namespace

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/statistic/info\?namespace=common
```


---


### /job
Jobs define a schedule where the core will provision runs for a pipeline.  

#### GET
Fetch the jobs present on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/job\?namespace=common
```

#### POST
Create a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/job -H "Content-Type: application/json" -d @docs/examples/jobs/hello-world.json
```

#### DELETE
Delete a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X DELETE -L http://127.0.0.1:8136/job\?id=hello-job
```


---


### /debug

#### GET
The endpoint shall be used to validate whether the core is online.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/debug
```

#### POST

##### Shutdown

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"shutdown\"}"
```

##### Latency

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"ping\"}"
```

##### Toggle Debug

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"debug\"}"
```




