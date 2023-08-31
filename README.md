# Mango ETL Batch & Stream Engine
An open source software orchestration framework for cloud functions implementing
the Extract Transform Load (ETL) data transformation pattern.

The main advantage of the ETLFramework is the ability for developers to independent functions focused on
data transformations that can behave like cloud functions but are deployed locally. The ETLFramework acts as
an orchestration engine responsible for exposing HTTP endpoints for on-demand funciton calling, scalability
around unexpected data loads, and performance monitoring for the functions developers write.


```tested on windows 10, macos, and ubuntu```

This framework is still a work in progress with hopes of adding more SRE functionality such as SLIs and SLAs
to the engine. If you are interested, feel free to reach out.

### Installation

```shell
   # create a local copy of the threads
   git clone https://github.com/GabeCordo/mango
   
   # generate a threads binary in the GOPATH bin folder
   go install
   
   # add $(go env GOPATH)/bin to your environment PATH
   
   # generate global files used by the threads when run
   mango init
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
