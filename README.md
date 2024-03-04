# Cluster.tools (aka. mango) Batch & Stream Engine

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main.svg?style=svg&circle-token=6bc46c7e268594646b3f38a6519d2209b7399ae2)](https://dl.circleci.com/status-badge/redirect/circleci/QC84aUAiJyQjmR73kpY2Vo/Vnh9fUxspVZXcLZeW3SfSR/tree/main)

An open source software orchestration framework for cloud functions implementing
the Extract Transform Load (ETL) data transformation pattern. 

The main advantage of the ETLFramework is the ability for developers to independent functions focused on
data transformations that can behave like cloud functions but are deployed locally. The ETLFramework acts as
an orchestration engine responsible for exposing HTTP endpoints for on-demand funciton calling, scalability
around unexpected data loads, and performance monitoring for the functions developers write.


```tested on windows 10, macos, and ubuntu```

This framework is still a work in progress with hopes of adding more SRE functionality such as SLIs and SLAs
to the engine. If you are interested, feel free to reach out.

### Local Installation

```shell
   # create a local copy of the threads
   git clone https://github.com/GabeCordo/mango
   
   # generate a threads binary in the GOPATH bin folder
   go install
   
   # add $(go env GOPATH)/bin to your environment PATH
   
   # generate global files used by the threads when run
   mango init
```

### Docker Installation

```shell
docker pull gabecordo/mango:main
docker exec gabecordo/mango:main mango help
```

### Running the ETL Engine

```shell
mango start
```
adding the common variant will load in util and test functions that can be used to verify the framework is working.

### Testing
Github actions are set to run all tests located in the repository. Component tests and Integration testing are used to
validate the health of the codebase.

Some tests are only suited for local runs; tests require special permissions or need some
form of manual intervention. To enable local tests, set the env var:

    MANGO_LOCAL_TEST

### Documentation

Documentation is continuously being added to the Github Wiki found [here](https://github.com/GabeCordo/mangoose-core/wiki)

### Disclosure

This repository is not related to the contributing members (of the repository) to the organizations they currently belong, the work they have, currently, or will perform at such organizations. All work completed within this repository pre-dates these organizations. All work completed withon this repository shall not be through company resources. Where "company resources" includes but is not limited to working hours, intellectual property, and electronic devices.

### Comercial Use

There are no limits on comercial use of this product. I highly discourage its use unless there is an intention of assisting with development in its current state. Bugs both seen-and-unseen do exist, and will likely be prevalent until extensive code coverage is completed. **If you are interested in using cluster.tools in your organization feel free to email me.**
