# controllers

Command Line Interface (CLI) controllers encapsulate functionality associated
with a unique identifier. For example, if the processes takes in the parameter 'start',
a start controller is defined such that anytime the identifier is invoked by an operator, the funciton
will be called.

## existing controllers

### doctor
Verify the required temporary files have been created and the general config used by PipelineOps is valid.

### init
Create the required temporary files and general config used by PipelineOps.

### logs
View a list of logs created by the PipelineOps process that exist in the local file system. 

### repl
Invoke an interactive shell to monitor the state of the PipelineOps process.

### schedule
Create a new execution schedule for a module/cluster pair. 

### start
Invoke the peppermint process.

### statistics
View a list of statistics created by the PipelineOps process that exist in the local file system.