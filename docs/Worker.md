# How to implement a Colonies Worker? 
The primary purpose of a worker is either to submit **process specifications** or **execute processes**. A process specification defines a process that should be executed in the future. Note that a Colonies process is generic concept and is not the same thing as an operating system process. A process is simply a series of computations or activities that produce some kind of result. It can be almost anything, for example turning on a lamp, training a neural network, serving a statistical model, dispatching a drone etc. 

A Colonies worker needs to implement a runtime environment in order to execute processes. To do so, it needs to interact with a Colonies server using the Colonies API. Every worker must have a valid runtime id and a corresponding runtime private key. The private key is used to sign all messages sent to the Colonies server. The server derives the runtime id from the signatures and then check if the worker is a member of the colony it tries to interact with.

## Submitting a process specification
A process specification can either be created by using the Colonies SDK available in Go, Python, Haskell, JavaScript or Julia, or by defining a JSON file as below and then submit it using the Colonies CLI. 

```json
{
  "conditions": {
    "runtimetype": "cli"
  },
  "func": "sleep",
  "args": [
    "3"
  ]
}
```

The condition attribute defines constraints or requirements on the workers. The **runtimetype** attribute defines which worker types are eligible to execute a process. Note that there can be many workers of the same runtime type. In this case, multiple workers *compete* executing processes. This is very useful for scaling computations beyond a single machine. It is also possible to directly specify an array of runtime IDs to more precisely control which worker should execute a process.   


The **func** attribute defines a function that should be executed by a worker. A worker might be capable of executing many functions. The **args** defines the arguments to the function. Note that it is up to the worker to interpret how to execute a particular process. After a worker has completed executing a function, it typically sets one or several output **attributes** on the process containing the result and then closes the process. As every process has an unique ID, other workers can then look up the process to retrieve the result. 

A process can either be closes as successful or failed. A process may automatically fail if some the following conditions are met:
1. A process has not been executed in a given time frame.
2. A process has executed too long time. 

```json
{
  "conditions": {
    "runtimetype": "cli"
  },
  "func": "sleep",
  "args": [
    "3"
  ],
  "maxwaittime": 10,
  "maxexectime": 5,
  "maxretries": 3
}
```

The **maxexectime** defines how many seconds a process may maximum execute before its moved back to queue maintained by the Colonies server. The **maxretries** specifies many times it may be moved to back to the queue before the process is automatically closed as a failure. These mechanisms are very useful for build fault tolerant systems. For example, a crashed worker will not be able to complete a process in time. In case, the process is automatically moved back to the queue so that other workers can execute it. The **maxretries** attribute prevents a process from bouncing around forever, for example if there is a bug in the worker code that prevents any worker from executing it, it will set a failed after the max retries has been reached. 

The **maxwaittime** defines how many seconds a process may be in the queue before it is automatically closed as a failure. This mechanism automatically cleans up the queue and let IT operation teams focus on investigating failed processes. If something is wrong, a process will eventually fail. It can also be useful if a process must be executed within a given time frame, for example a user may have a requirement that a lamp must be turned on within a second, or something is wrong. 

In the JSON example above, the sleep process must be completed in 5 seconds. This is ok since it will only sleep for 3 seconds. However, if we change the sleep args to 6 seconds, the worker will get an error message when it closes the process since it has timed out. As it is impossible in this case to complete the process in time, it will go back to the queue 3 times before it is finally closed as failed. The process will also fail if a worker has not been assigned the process within 10 seconds. 

##  
```json
{
  "conditions": {
    "runtimetype": "cli"
  },
  "func": "sleep",
  "args": [
    "3"
  ],
  "env": {
    "TEST": "testenv"
  }
}
```

It is also possible to specify environment variables (key-value pairs) as a complement to the args attribute. The **env** dictionary is automatically converted to attributes by the Colonies server, which can then be retrieved by the worker code after assignment. When using the built-in worker CLI (colonies worker), the env dictionary is automatically converted to standard OS environmental variables.   

## Implementing a worker
A worker connects to the Colonies server and tries to assign a process. This is done by sending an assign request. Note that a worker is not guaranteed to get a process. There are several reasons why an assign request may fail. 

1. The queue is empty or and another competing worker was assigned the process instead. 
2. The Colonies server is temporary down. This may for example happen during an upgrade, or if a Colonies server Kubernetes instance (Pod) has been removed, e.g. by a chaos monkey or system failure.  

In case of an error, the worker should keep calling the assign method until it is assigned a process. 

Golang worker example:
```go
for {
    process, err := client.AssignProcess(colonyID, timeout, prvKey)
    if err!=nil {
        continue
    }
    
    execute(process)
    client.CloseSuccessful(process.ID, prvKey)
}
```

Note the **timeout** argument. The worker must specify how long time it is willing to wait for process. That is, how long time the **AssignProcess** function should block. The timeout should be set to at least 1 second to prevent overloading the Colonies with too many assign requests.

Also note that there is no guarantee that the AssignProcess function actually returns a process even if the function has not timed out. Another worker might have been quicker and was assigned the process.

Julia worker example:
```julia
    while true
        try
            process = ColonyRuntime.assignprocess(client, timeout, colonyid, prvkey)
            execute(process)
            ColonyRuntime.closeprocess(client, process.processid, true, prvkey)
        catch err
            # ignore, just re-try
        end
    end
end
```

Javascript worker example: 
```javascript
      function assign() {
          runtime.assign(colonyid, runtime_prvkey).then((process, err) => {
              if err !== undefined {
                  execute(process)
                  runtime.close_process(process.processid, true, prvkey)
              }
          }) 
      } 

      runtime.load().then(() => {
        runtime.subscribe_processes(runtime_type, 0, prvkey, function(processes) {
           assign()
        })
      });
```

In Javascript it might be useful to use Colonies pubsub websocket protocol to avoid blocking the browser main thread.

# Python worker example:
```python
    while True:
       try
           process = client.assign_process(colonyid, timeout, prvkey)
           execute(process)
           client.close_process(process.processid, prvkey)
        except: 
           pass  # just ignore
       
```


## Service discovery
As Colony contains all registered workers (runtimes), it is possible to use it for service discovery, e.g. search for a particular worker and submit a process specification directly to it.   

```go
    runtimes, err := client.GetRuntime(colonyID, prvKey)
    for _, runtime := range runtimes {
        if runtime.Name == "videocam" {
             condition := &Condition{RuntimeID: []{runtime.ID}, ColonyID: colonyID}
             processSpec := &ProcessSpec{Condition: condition, Func: "turn_on_video", Args: []{arg}, MaxExecTime: 1, MaxRetries: 3}
             err := client.SubmitProcessSpec(processSpec, prvKey)
        }
    }
}
```
