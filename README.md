# ETLFramework

### CLI Parameters

#### -h --help
Provides descriptions for available parameters.

#### -d --debug
Provides verbose output for the starter script.

#### -g --generate-key
ECDSA public and private keys are outputted as x509 encoded formats.

---

### HTTP Curl Interaction

##### Shutdown Node
curl -X GET http://127.0.0.1:8000/debug -H 'Content-Type: application/json' -d '{"function": "shutdown"}'

##### Test Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "multiply"}'

##### Test Cluster Statistics
curl -X GET http://127.0.0.1:8000/statistics -H 'Content-Type: application/json' -d '{"function": "multiply"}'