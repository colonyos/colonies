# Tutorial
In this tutorial we will develop a Colonies Executor using the Golang SDK. The executor calculates the last number in a given Fibonacci serie.

## 1. Set up a Colonies development server
```console
colonies dev
```
## 2. Environmental variables
```console
source examples/devenv
}
```

## 3. Fibonacci Job Generator code (examples/generator.go)
```go
func main() {
    colonyID := os.Getenv("COLONIES_COLONY_ID")
    executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")
    coloniesHost := os.Getenv("COLONIES_SERVER_HOST")
    coloniesPortStr := os.Getenv("COLONIES_SERVER_PORT")
    coloniesPort, err := strconv.Atoi(coloniesPortStr)
    if err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }

    funcSpec := core.CreateEmptyFunctionSpec()
    funcSpec.Conditions.ColonyName = colonyID
    funcSpec.Conditions.ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
    funcSpec.Env["fibonacciNum"] = os.Args[1]

    client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
    addedProcess, err := client.Submit(funcSpec, executorPrvKey)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println("Submitted a new process to the Colonies server with Id <" + addedProcess.ID + ">")
}
```


## 6. Fibonacci Solver executor code (examples/solver.go) 
```go
func main() {
    colonyID := os.Getenv("COLONIES_COLONY_ID")
    executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")
    coloniesHost := os.Getenv("COLONIES_SERVER_HOST")
    coloniesPortStr := os.Getenv("COLONIES_SERVER_PORT")
    coloniesPort, err := strconv.Atoi(coloniesPortStr)
    if err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }

    // Ask the Colonies server to assign a process to this executor 
    client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
    assignedProcess, err := client.Assign(colonyID, 100, executorPrvKey) // Max wait 100 seconds for assignment request
    if err != nil {
        fmt.Println(err)
        return
    }

    // Parse env attribute and calculate the given Fibonacci number
    for _, attribute := range assignedProcess.Attributes {
        if attribute.Key == "fibonacciNum" {
            fmt.Println("We were assigned process " + assignedProcess.ID)
            fmt.Println("Calculating Fibonacci serie for " + attribute.Value)
            nr, _ := strconv.Atoi(attribute.Value)
            fibonacci := fib.FibonacciBig(uint(nr))
            fmt.Println("Result: The last number in the Fibonacci serie " + attribute.Value + " is " + fibonacci.String())

            attribute := core.CreateAttribute(assignedProcess.ID, colonyID, "", core.OUT, "result", fibonacci.String())
            client.AddAttribute(attribute, executorPrvKey)

            // Close the process as successful
            client.Close(assignedProcess.ID, executorPrvKey)
            return
        }
    }

    // Close the process as failed
    client.Fail(assignedProcess.ID, "invalid args", executorPrvKey)
}
```

## 7. Calculating Fibonacci numbers
### Generate a job
```console
go run generator.go 1234 
```

Output:
```
Submitted a new process to the Colonies server with Id <4c19d59be7ad02d27491c993d7deaff4f58ffad55bbddb7200fb638299820da4>
```

### Look up the job in queue 
```console
colonies process psw --insecure
```

```console
+------------------------------------------------------------------+-----+------+---------------------+---------------+
|                                ID                                | CMD | ARGS |   SUBMISSION TIME   | EXECUTOR TYPE |
+------------------------------------------------------------------+-----+------+---------------------+---------------+
| 705abd98cb2f801aa4c0a357c367ea8a5cc89a51d24aaadbca89abbb4be00b7e |     |      | 2022-05-27 14:10:12 | cli           |
+------------------------------------------------------------------+-----+------+---------------------+---------------+
```

### Run one job from the queue 
```console
go run solver.go
```

Output:
```
We were assigned process 705abd98cb2f801aa4c0a357c367ea8a5cc89a51d24aaadbca89abbb4be00b7e
Calculating Fibonacci serie for 12
Result: The last number in the Fibonacci serie 12 is 144
```

```console
colonies process get --processid 705abd98cb2f801aa4c0a357c367ea8a5cc89a51d24aaadbca89abbb4be00b7e --insecure
```

```
Process:
+---------------------+------------------------------------------------------------------+
| ID                  | 705abd98cb2f801aa4c0a357c367ea8a5cc89a51d24aaadbca89abbb4be00b7e |
| IsAssigned          | True                                                             |
| AssignedExecutorID  | 3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac |
| State               | Successful                                                       |
| SubmissionTime      | 2022-05-27 14:10:03                                              |
| StartTime           | 2022-05-27 14:12:41                                              |
| EndTime             | 2022-05-27 14:12:41                                              |
| Deadline            | 0001-01-01 01:12:12                                              |
| WaitingTime         | 2m37.735526s                                                     |
| ProcessingTime      | 9.11ms                                                           |
| Retries             | 0                                                                |
+---------------------+------------------------------------------------------------------+

FunctionSpec:
+-------------+------+
| Func        | None |
| Args        | None |
| MaxExecTime | -1   |
| MaxRetries  | 3    |
+-------------+------+

Conditions:
+--------------+------------------------------------------------------------------+
| ColonyName     | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| ExecutorIDs  | None                                                             |
| ExecutorType | cli                                                              |
| Memory       | 1000                                                             |
| CPU Cores    | 10                                                               |
| GPUs         | 1                                                                |
+--------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+--------------+-------+------+
|                                ID                                |     KEY      | VALUE | TYPE |
+------------------------------------------------------------------+--------------+-------+------+
| c288d631ae86efc84c54b4c40e2420845d9ac04aecfa614d30f2a509441994b2 | fibonacciNum | 12    | Env  |
| 798040fccd6100fd68f680cdd962c87caf5098e826797c1e85b154dbecf87a27 | result       | 144   | Out  |
+------------------------------------------------------------------+--------------+-------+------+
```

See examples/generate_sub.go and examples/solver_pub.go for an event-driven version of the generator and executor.
