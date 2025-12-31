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

### 2.2. /module

#### 2.2.1. GET
Fetch a set of [module](#12-module) records from the core.

#### 2.2.2. PUT

### 2.3 /function
Fetch a set of [function](#13-function) records from the core.

#### 2.3.1. GET

#### 2.3.2. PUT

### 2.4. /namespaces

#### 2.4.1. GET
Fetch a set of namespaces used in the core. Namespaces are used to allow scoped use of pipeline and job identifiers rather
than maintaining a global namespace.

The _common_ namespace is the default global namespace that is present in all cores.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/namespaces
```
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

#### 2.5.3. PUT

#### DELETE

### 2.6. /run

#### 2.6.1. GET

###### HTTP Params
namespace (mandatory)

pipeline (optional)

maximumResults (optional)

offsetOfResults (optional)

#### 2.6.2. POST

### 2.7. /statistics

#### 2.7.1. GET

### 2.8. /statistic

#### 2.8.1. GET

### 2.9. /job

#### 2.9.1. GET

#### 2.9.2. POST

#### 2.9.3. DELETE

### 2.10. /debug

#### 2.10.1. GET
The endpoint shall be used to validate whether the core is online.

#### 2.10.2. POST

##### Shutdown

###### HTTP Body
```json
{
  "action": "shutdown"
}
```

##### Latency

###### HTTP Body
```json
{
  "action": "ping"
}
```

##### Toggle Debug

###### HTTP Body
```json
{
  "action": "debug"
}
```




