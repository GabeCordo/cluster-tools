# PipelineOps (POPS) Framework

An open source ingress gateway for monitoring and load balancing requests to distributed data pipelines. 

Pops provides fine-grained control over a deployment of horizontally distribute pipelines to control: how they
are provisioned, how they should be taken offline, how they should be made redundant, and how they should be versioned.

> The gateway is still a work in progress with hopes of adding more SRE functionality such as SLIs and SLAs
> to the engine. If you are interested, feel free to reach out.

### Local Installation
Before performing a local installation make sure the GOPATH bin folder has been added to your environment PATH variable. The
'go install' command is a quick way to build and store a binary inside $(go env GOPATH)/bin. You will not be able to call a binary
installed with 'go install' otherwise.

```shell
   # create a log copy of the thread
   git clone https://github.com/Sentmint/PipelineOps
   
   # install the pipeline-gateway binary
   cd /cmd/pipeline-gateway
   # generate the pipeline-gateway binary in the GOPATH bin folder
   go install
   # generate global files used by the thread when statistic
   pipeline-gateway init
   # validate pipeline-gateway is installed correctly
   pipeline-gateway doctor
   
   # install the pops binary
   cd ../pops
   # generate the pops binary in the GOPATH bin folder
   go install
  
```

### Running the POPS Gateway
The gateway is an orchestrator that manages various pipeline deployments. The developer communicates with the gateway to
create, run, and watch pipelines defined by yaml files.

```shell
pipeline-gateway start
```

### Testing
Test are being migrated from Github Actions to CircleCI. Component tests and Integration testing are used to
validate the health of the codebase.

Code Coverage improvements are underway to increase confidence in code correctness. We believe maintaining a reasonable level
of code coverage is a crucial step in convincing individuals to try PipelineOps as an infrastructure solution.

### Documentation

Documentation can be found inside the docs folder [here](docs).
