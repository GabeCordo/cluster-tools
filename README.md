# ETLFramework

## Test Functions

##### Shutdown Node
curl -X GET http://127.0.0.1:8000/debug -H 'Content-Type: application/json' -d '{"function": "shutdown"}'

##### Multiply Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "multiply"}'