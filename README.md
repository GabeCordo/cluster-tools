# Cluster.tools Cloud Framework

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main.svg?style=svg&circle-token=6bc46c7e268594646b3f38a6519d2209b7399ae2)](https://dl.circleci.com/status-badge/redirect/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main) [![codecov](https://codecov.io/gh/GabeCordo/cluster-tools/graph/badge.svg?token=OCLP5E8E4J)](https://codecov.io/gh/GabeCordo/cluster-tools)

An open source ingress gateway for monitoring and load balancing requests to distributed data pipelines. 

Cluster.tools provides fine-grained control over a deployment of horizontally distribute pipelines to control: how they
are provisioned, how they should be taken offline, how they should be made redundant, and how they should be versioned.

> The gateway is still a work in progress with hopes of adding more SRE functionality such as SLIs and SLAs
> to the engine. If you are interested, feel free to reach out.

### Local Installation

```shell
   # create a log copy of the thread
   git clone https://github.com/GabeCordo/cluster-tools
   
   # generate a thread binary in the GOPATH bin folder
   go install
   
   # add $(go env GOPATH)/bin to your environment PATH
   
   # generate global files used by the thread when run
   cluster-tools init
   
   # validate cluster-tools installed correctly
   cluster-tools doctor
```

### Running the Cluster.tools Process

```shell
cluster-tools start
```

### Testing
Test are being migrated from Github Actions to CircleCI. Component tests and Integration testing are used to
validate the health of the codebase.

Code Coverage improvements are underway to increase confidence in code correctness. We believe maintaining a reasonable level
of code coverage is a crucial step in convincing individuals to try cluster.tools as an infrastructure solution.

### Documentation

Documentation is continuously being added to the Github Wiki found [here](https://cluster.tools)

### Commercial Use

Anyone is free to use cluster.tools inside their production environments **but it is not unlikely this comes with the
raising fixes until the project becomes mature.**

### Disclosure

This repository is not related to the contributing members (of the repository) to the organizations they currently belong, the work they have, currently, or will perform at such organizations. All work completed within this repository pre-dates these organizations. All work completed withon this repository shall not be through company resources. Where "company resources" includes but is not limited to working hours, intellectual property, and electronic devices.
