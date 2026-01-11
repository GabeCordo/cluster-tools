# core architecture decisions

## thread arrangement

![](/.bin/images/threads_diagram.png)

## procedures

### Processor Connects to Core

```mermaid
sequenceDiagram
    socket ->> processor : AddProcessor;
    break processor already exists
        processor ->> socket : Fail ProcessorAlreadyExists
    end
    processor ->> processor : Create Processor Record
    processor ->> socket : Response;
```

### Processor Disconnects from Core

```mermaid
sequenceDiagram
    socket ->> processor : DeleteProcessor;
    break processor already exists
        processor ->> socket : Fail ProcessorDoesNotExist
    end
    processor ->> processor : Delete Processor Record
    processor ->> socket : Response;
```

### Processor Registers Module

```mermaid
sequenceDiagram
    socket ->> processor : AddModule;
    break when processor does not exist
        processor ->> socket : Fail ProcessorDoesNotExist
    end
    processor ->> processor : Create Module Record
    processor ->> socket : Response;
```

### Create Pipeline

```mermaid
sequenceDiagram
    rest ->> database : AddPipeline
    database ->> rest : Response
```

### Start Run

```mermaid
sequenceDiagram
    autonumber
    actor operator;
    operator ->> rest : POST /run;
    rest ->> processor : RunPipeline;
    processor ->> database : GetPipeline;
    database ->> processor : Response;
    processor ->> processor : FindCandidateProcessor;
    processor ->> runner : CreateRun;
    runner ->> database : GetPipeline;
    database ->> runner : Response;
    runner ->> runner : CreateRunRecord;
    runner ->> socket : CreateRun;
    socket -->> processor_service : Message(Create, Run)
    socket ->> runner : Response;
    runner ->> runner : Set Run Status;
    runner ->> processor : Response;
    processor ->> rest : Response;
    rest ->> operator : 200 OK;
```

### Stop Run

```mermaid
sequenceDiagram
    autonumber
    actor operator;
    operator ->> rest : DELETE /run?id
    rest ->> processor : StopRun
    processor ->> runner : StopRun
    break run does not exist
        runner ->> processor : Response(Fail)
        processor ->> rest : Response(Fail)
        rest ->> operator : 404 NOT FOUND
    end
    runner ->> socket : StopRun
    socket ->> processor_service : Message(Stop, Run)
    socket ->> runner : Response
    runner ->> processor : Response
    processor ->> rest : Response
    rest ->> operator : 200 OK
```

### Update Run

```mermaid
sequenceDiagram
    autonumber
    socket ->> processor : UpdateRun
    processor ->> runner : UpdateRun
    break run does not exist
        runner ->> processor : Error Response
        processor ->> socket : Error Response
    end
    runner ->> runner : Set Statistics
    alt runner has completed or crashed
        runner ->> database : Create Statistic
        database ->> runner : Response
        runner ->> messenger : Close Messenger For Run
        messenger ->> messenger : Flush Logs for Run
        
    end
```