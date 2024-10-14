# Cluster.tools Cloud Framework

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main.svg?style=svg&circle-token=6bc46c7e268594646b3f38a6519d2209b7399ae2)](https://dl.circleci.com/status-badge/redirect/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main) [![codecov](https://codecov.io/gh/GabeCordo/cluster-tools/graph/badge.svg?token=OCLP5E8E4J)](https://codecov.io/gh/GabeCordo/cluster-tools)

An open source ingress gateway for monitoring and load balancing requests to distributed data pipelines. 

Cluster.tools provides fine-grained control over a deployment of horizontally distribute pipelines to control: how they
are provisioned, how they should be taken offline, how they should be made redundant, and how they should be versioned.

> The gateway is still a work in progress with hopes of adding more SRE functionality such as SLIs and SLAs
> to the engine. If you are interested, feel free to reach out.

### Local Installation
Before performing a local installation make sure the GOPATH bin folder has been added to your environment PATH variable. The
'go install' command is a quick way to build and store a binary inside $(go env GOPATH)/bin. You will not be able to call a binary
installed with 'go install' otherwise.

```shell
   # create a log copy of the thread
   git clone https://github.com/GabeCordo/cluster-tools
   
   # install the ctgate binary
   cd /cmd/ctgate
   # generate the ctgate binary in the GOPATH bin folder
   go install
   # generate global files used by the thread when statistic
   ctgate init
   # validate ctgate is installed correctly
   ctgate doctor
   
   # install the ctools binary
   cd ../ctools
   # generate the ctools binary in the GOPATH bin folder
   go install
  
```

### Running the Cluster.tools Gateway
The gateway is an orchestrator that manages various pipeline deployments. The developer communicates with the gateway to
create, run, and watch pipelines defined by yaml files.

```shell
ctgate start
```

### Testing
Test are being migrated from Github Actions to CircleCI. Component tests and Integration testing are used to
validate the health of the codebase.

Code Coverage improvements are underway to increase confidence in code correctness. We believe maintaining a reasonable level
of code coverage is a crucial step in convincing individuals to try cluster.tools as an infrastructure solution.

### Documentation

Documentation can be found inside the docs folder [here](docs).
