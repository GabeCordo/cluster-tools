# Extract Transform Load Engine
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
   git clone https://github.com/GabeCordo/etl
   cd etl
   mkdir .bin/modules
   # add .bin/modules as to your environment as ETL_ENGINE_MODULES
   # add .bin/configs/config.etl.yaml to your environment as ETL_ENGINE_CONFIG
   go install
   # add $(go env GOPATH)/bin to your environment PATH
```

### Running the ETL Engine

```shell
etl --config $ETL_ENGINE_CONFIG --modules ETL_ENGINE_MODULES
```

### Documentation

Documentation is continuously being added to the Github Wiki found [here](https://github.com/GabeCordo/etl/wiki)