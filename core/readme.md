# core
The core is a collection of source files required by the cluster-tools process.

## Folders

### api
Functions responsible for communicating with external processors.

### components
Functions that contain synchronous logic that is invoked by threads using asynchronous message passing.

For example, if we want to create a database responsible for storing configurations. The database is a synchronous component whos
functions are invoked asynchronously by requests destined to a thread.

### interfaces
Data structures used by the core or external processors. 

### threads
Functions that contain asynchronous functionality responsible for handling incoming requests
that invoke a synchronous component.