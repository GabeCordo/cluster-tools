# core



---

## 1. Models

### 1.1. Processor

### 1.2. Module

### 1.3. Function

### 1.4. Pipeline

### 1.5. Run

### 1.6. Statistic



---



## 2.0 HTTP APIs
Defines the endpoints used to query information inside the core.

### 2.1 /processor

#### 2.1.1. GET
Fetch a set of [processor](#11-processor) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/processor
```

---


### 2.2. /module

#### 2.2.1. GET
Fetch a set of [module](#12-module) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/module
```

#### 2.2.2. PUT


---


### 2.3 /function
Fetch a set of [function](#13-function) records from the core.

#### 2.3.1. GET

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/function\?module=common
```

#### 2.3.2. PUT


---


### 2.4. /namespaces

#### 2.4.1. GET
Fetch a set of namespaces used in the core. Namespaces are used to allow scoped use of pipeline and job identifiers rather
than maintaining a global namespace.

The _common_ namespace is the default global namespace that is present in all cores.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/namespaces
```


---


### 2.5. /pipeline

#### 2.5.1. GET

###### HTTP Params
namespace 

pipeline _(optional)_

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/pipeline\?namespace=common
```

#### 2.5.2. POST
Create a new pipeline that defines the flow of data between functions run on processors.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/pipeline -H "Content-Type: application/json" -d @docs/examples/pipelines/hello-world.json
```

#### 2.5.3. PUT
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


### 2.6. /run

#### 2.6.1. GET
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

#### 2.6.2. POST
Provision a new running instance of a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/run -H "Content-Type: application/json" -d @docs/examples/runs/hello-world.json
```

### 2.7 /run/count

#### 2.7.1. GET
Get the number of runs for a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/count\?namespace=common\&pipeline=hello-world
```

---


### 2.8. /statistic

#### 2.8.1. GET
Fetch the statistics collected from completed pipeline runs.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/statistic\?namespace=common\&pipeline=hello-world
```


---


### 2.9. /statistic/info

#### 2.9.1. GET

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


### 2.10. /job
Jobs define a schedule where the core will provision runs for a pipeline.  

#### 2.10.1. GET
Fetch the jobs present on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/job\?namespace=common
```

#### 2.10.2. POST
Create a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/job -H "Content-Type: application/json" -d @docs/examples/jobs/hello-world.json
```

#### 2.10.3. DELETE
Delete a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X DELETE -L http://127.0.0.1:8136/job\?id=hello-job
```


---


### 2.11. /debug

#### 2.11.1. GET
The endpoint shall be used to validate whether the core is online.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/debug
```

#### 2.11.2. POST

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




