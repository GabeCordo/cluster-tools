# controllers

Command Line Interface (CLI) controllers encapsulate functionality associated
with a unique identifier. For example, if the processes takes in the parameter 'start',
a start controller is defined such that anytime the identifier is invoked by an operator, the funciton
will be called.

## existing controllers

### doctor
Verify the required temporary files have been created and the general config used by cluster-tools is valid.

### init
Create the required temporary files and general config used by cluster-tools.

### logs
View a list of logs created by the cluster-tools process that exist in the local file system. 

### repl
Invoke an interactive shell to monitor the state of the cluster-tools process.

### schedule
Create a new execution schedule for a module/cluster pair. 

### start
Invoke the cluster.tools process.

### statistics
View a list of statistics created by the cluster-tools process that exist in the local file system.