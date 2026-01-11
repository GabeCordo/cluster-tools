# core-processor tunnel
This specification covers the TCP tunnel between the **core** and **processor** services.

You can implement the tunnel protocol to create custom **processor** services and communicate with the core.

---

## Tunnel Message

### Message Types

###### Message Action (uint8)
What type of functionality is invoked by the message.

1. Ping (0)
2. Create (1)
3. Update (2)
4. Delete (3)

###### Message Type (uint8)
The type of data to decode from the message.

1. Module (0)
2. Run (1)
3. Log (2)

###### Message Data (structure)
The payload of the message.

### Message Payloads

###### Module
The module structure carries information about the runnable functions available on the processor.

```text
module => {
    id : string
    version : string
    contact => {
        name : string
        email : string
    }
    functions => []{
        id : string
        parameters => []string
        returns => []string
    }
}
```

###### Run
The run structure carries information about the running function(s) on the prcessor.

```text
run => {
    id : uint64
    status => enum{
        created
        activated
        crashed
        completed
        terminated
        cancelled
    }
    time => {
        created : time (optional)
        last_updated : time (optional)
        duration => {
            hours : uint64
            minutes : uint64
            seconds : uint64
            milliseconds : uint64
        } (optional)
        started_by : => enum{
            operator
            processor
            scheduler
        }
        processor : uint64 (optional)
        namespace : string (optional)
        pipeline => {
            identifier : string
            on_crash => enum {
                Restart
                DoNothing
            }
            functions => []{
                module : string
                identifier : string
                metadata => {
                    static_mount : bool
                }
                from : string (optional)
                to : string (optional)
                start_with : uint16 (optional)
                wait_before : bool (optional)
                maximum : uint16 (optional)
                parameters => []string
                returns => []string
            }
            pipes => []{
                identifier : string
                threshold : uint32
                growth_factor : float64
            }
        } (optional)
        statistics => {
            num_of_functions : uint16
            functions => []{
                active : uint16
                provisions : uint16
            }
            num_of_pipes : uint16
            pipes => []{
                pushed : uint64
                pulled : uint64
                dropped : uint64
                breaches : uint64
                timing => {
                    min_time_before_pop_ns : uint64
                    max_time_before_pop_ns : uint64
                    average_time_ns : uint64
                    median_time_ns : uint64
                }
            }
        } (optional)
    }
}
```

###### Log
Log has not been designed.

### Message Format
The message sent between the services is formated as such. The _type_ field specifies how the service should decode the 
content of the _data_ field.

```text
message => {
    action : MessageAction
    type : MessageType
    data : Module | Run | Log
}
```

---

## Tunnels Events
Events sent across the tunnel.

### Events Initiated by the Processor
Events sent from the **Processor** to **Core** services on the TCP tunnel.

#### Processor Connect
The processor sends a message to the core informing it that the service has come online. 

The core cannot perform
any actions on the processor until the processor sends a module of runnable functions.

###### Message Content
```json
{
  
}
```

#### Processor Disconnect


#### Processor Create Module


#### Processor Update Run


### Events Initiated by the Core
Events sent from the **Core** to **Processor** services on the TCP channel.

#### Core Create Run


#### Core Stops Run


---