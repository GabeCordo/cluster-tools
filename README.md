# Flock

An open source ingress gateway for monitoring and load balancing requests to distributed data pipelines. 

Pops provides fine-grained control over a deployment of horizontally distribute pipelines to control: how they
are provisioned, how they should be taken offline, how they should be made redundant, and how they should be versioned.

> Flock is a personal project that is a work in progress. 

### Local Flock Installation
Before performing a local installation make sure the GOPATH bin folder has been added to your environment PATH variable. The
'go install' command is a quick way to build and store a binary inside $(go env GOPATH)/bin. You will not be able to call a binary
installed with 'go install' otherwise.

```shell
   # create a log copy of the thread
   git clone https://github.com/FortifiedCode/flock
   
   # install the flock binary
   cd /cmd/flock
   # generate the flock binary in the GOPATH bin folder
   go install
   # generate global files used by the thread when statistic
   flock init
   # validate flock is installed correctly
   flock doctor
  
```

### Running Flock
Flock is an orchestrator that manages pipeline operations.

```shell
flock start
```

### Testing
Unit Tests and Code Coverage is a priority for future areas of work.

### Documentation

Documentation can be found inside the docs folder [here](docs/readme.md).
