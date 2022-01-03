# Tutorial
In this we will develop two Runtimes in Golang. The first Runtime will generate pending process specifications containing an integer. The other Runtime will fetch these specifications and start the process where it calculates a Fibonacci number of the integer.

## 1. Set up a Colonies server [(see intructions here)](./Installation.md)

## 2. Register a new Colony
```console
./bin/colonies colony register --serverid=9289dfccedf27392810b96968535530bb69f90afe7c35738e0e627f3810d943e --spec ./examples/colony.json
```
Output:
```
50d74cb4c8306856a4c854089280c6be80353b36e6f61c41c435f8c87c9ec1cb
```
The output is the ColonyID, we'll need it later.

```console
export COLONYID="50d74cb4c8306856a4c854089280c6be80353b36e6f61c41c435f8c87c9ec1cb"
```

## 3. Register a Runtime for Fibonacci Job Generator
```json
{
    "name": "FibonacciJobGenerator",
    "runtimetype": "FibonacciJobGenerator",
    "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
    "cores": 32,
    "mem": 80326,
    "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
    "gpus": 1
}
```

```console
./bin/colonies runtime register --spec examples/fibonacci/generator.json
```

Output:
```
94227a8a9bbe6459de6d83414083300408066ef29eb179845219edb1e7349ccc
```

## 3. Register a Runtime for Fibonacci solver 
```json
{
    "name": "FibonacciSolver",
    "runtimetype": "FibonacciSolver",
    "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
    "cores": 32,
    "mem": 80326,
    "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
    "gpus": 1
}
```

```console
./bin/colonies runtime register --spec examples/fibonacci/solver.json
```

Output:
```
964cb72ef11deab2f66f089fdbd5d4b5df934af0774aa67fb0695e196df4d1a8
```

## 4. Approve the Rutimes
```console
./bin/colonies runtime approve --runtimeid 94227a8a9bbe6459de6d83414083300408066ef29eb179845219edb1e7349ccc
./bin/colonies runtime approve --runtimeid 964cb72ef11deab2f66f089fdbd5d4b5df934af0774aa67fb0695e196df4d1a8
```

## 5. Look up the Private Key of the two Runtimes
We will need the ColonyID and Private Key of the two Runtimes in the Go-code below.

### Fibonacci Job Generator
```console
./bin/colonies keychain get --id 94227a8a9bbe6459de6d83414083300408066ef29eb179845219edb1e7349ccc
```
Output:
```
a7c4bd22e94010027b31bf669c766aed0e24d6e4e1da05511b56440238df108f
```

### Fibonacci Solver
```console
./bin/colonies keychain get --id 964cb72ef11deab2f66f089fdbd5d4b5df934af0774aa67fb0695e196df4d1a8
```
Output:
```
dc99f66eb882813916135fc0f1d913c38b2ac0435f3e3fae60eb70b421a92e28
```

## 6. Fibonacci Job Generator code (generator.go)
```go
package main

import (
    "colonies/pkg/client"
    "colonies/pkg/core"
    "fmt"
    "os"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]
    fibonacciNum := os.Args[2]

    env := make(map[string]string)
    env["fibonacciNum"] = fibonacciNum

    processSpec := core.CreateProcessSpec(colonyID, []string{}, "FibonacciSolver", -1, 3, 1000, 10, 1, env)

    client := client.CreateColoniesClient("localhost", 8080, true)
    _, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
    }
}
```


## 6. Fibonacci Solver code (solver.go) 
```go
package main

import (
    "colonies/pkg/client"
    "fmt"
    "os"
    "strconv"

    fib "github.com/t-pwk/go-fibonacci"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]

    // Ask the Colonies server to assign a process to this Runtime
    client := client.CreateColoniesClient("localhost", 8080, true)
    assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Parse env attribute and calculate the given Fibonacci number
    for _, attribute := range assignedProcess.Attributes {
        if attribute.Key == "fibonacciNum" {
            nr, _ := strconv.Atoi(attribute.Value)
            fmt.Println(fib.FibonacciBig(uint(nr)))

            // Close the process as Successful
            client.MarkSuccessful(assignedProcess.ID, runtimePrvKey)
            return
        }
    }

    // Close the process as Failed
    client.MarkFailed(assignedProcess.ID, runtimePrvKey)
}
```

## 7. Submit a task to the solver
```console
go run generator.go a7c4bd22e94010027b31bf669c766aed0e24d6e4e1da05511b56440238df108f 1234 
```

```console
go run solver.go dc99f66eb882813916135fc0f1d913c38b2ac0435f3e3fae60eb70b421a92e28
```

```
Output:
347746739180370201052517440604335969788684934927843710657352239304121649686845967975636459392453053377493026875020744760145842401792378749321113719919618588095724485583919541019961884523908359133457357334538791778480910430756107407761555218113998374287548487
```

## 8. Improved solver (solver_sub.go)
In the example above, we need to manually start a new solver to fetch a pending process. What about if we could subscribe for events when there are new processes available?

This can be done with **SubscribeProcesses** function below:
```go
subscription, err := client.SubscribeProcesses("FibonacciSolver", core.WAITING, 100, runtimePrvKey)
```

```go
package main

import (
    "colonies/pkg/client"
    "colonies/pkg/core"
    "fmt"
    "os"
    "strconv"

    fib "github.com/t-pwk/go-fibonacci"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]

    // Ask the Colonies server to assign a process to this Runtime
    client := client.CreateColoniesClient("localhost", 8080, true)

    subscription, err := client.SubscribeProcesses("FibonacciSolver", core.WAITING, 100, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
        return
    }

    go func() {
        for {
            select {
            case <-subscription.ProcessChan:
                assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
                if err != nil {
                    fmt.Println(err)
                    continue
                }

                // Parse env attribute and calculate the given Fibonacci number
                for _, attribute := range assignedProcess.Attributes {
                    if attribute.Key == "fibonacciNum" {
                        nr, _ := strconv.Atoi(attribute.Value)
                        fmt.Println(fib.FibonacciBig(uint(nr)))

                        // Close the process as Successful
                        client.MarkSuccessful(assignedProcess.ID, runtimePrvKey)
                        continue
                    }
                }

                // Close the process as Failed
                client.MarkFailed(assignedProcess.ID, runtimePrvKey)
            case err := <-subscription.ErrChan:
                fmt.Println(err)
            }
        }
    }()

    // Wait forever
    <-make(chan bool)
}
```

In one terminal type:
```console
go run solver_sub.go dc99f66eb882813916135fc0f1d913c38b2ac0435f3e3fae60eb70b421a92e28
```

In another terminal type: 
```console
go run generator.go a7c4bd22e94010027b31bf669c766aed0e24d6e4e1da05511b56440238df108f 43
```

After typing the command, a Fibonacci will appear in the first terminal, e.g.

```console
go run solver_sub.go dc99f66eb882813916135fc0f1d913c38b2ac0435f3e3fae60eb70b421a92e28
347746739180370201052517440604335969788684934927843710657352239304121649686845967975636459392453053377493026875020744760145842401792378749321113719919618588095724485583919541019961884523908359133457357334538791778480910430756107407761555218113998374287548487
347746739180370201052517440604335969788684934927843710657352239304121649686845967975636459392453053377493026875020744760145842401792378749321113719919618588095724485583919541019961884523908359133457357334538791778480910430756107407761555218113998374287548487
```

## 9. Sending result back to the generator
Let's first improve the solver code to save Fibonacci number as an Attribute before closing the process.

This can be done with this additional code:
```go
attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", fibonacci.String())
client.AddAttribute(attribute, runtimePrvKey)
```

### Improved solver (solver_ret.go) 
```go
package main

import (
    "colonies/pkg/client"
    "colonies/pkg/core"
    "fmt"
    "os"
    "strconv"

    fib "github.com/t-pwk/go-fibonacci"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]

    // Ask the Colonies server to assign a process to this Runtime
    client := client.CreateColoniesClient("localhost", 8080, true)
    assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Parse env attribute and calculate the given Fibonacci number
    for _, attribute := range assignedProcess.Attributes {
        if attribute.Key == "fibonacciNum" {
            nr, _ := strconv.Atoi(attribute.Value)
            fibonacci := fib.FibonacciBig(uint(nr))

            attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", fibonacci.String())
            client.AddAttribute(attribute, runtimePrvKey)

            // Close the process as Successful
            client.MarkSuccessful(assignedProcess.ID, runtimePrvKey)
            return
        }
    }

    // Close the process as Failed
    client.MarkFailed(assignedProcess.ID, runtimePrvKey)
}
```

Let's improve the generator. First improvement is just to print process ID. 

### Improved generator (generator2.go)
```go
package main

import (
    "colonies/pkg/client"
    "colonies/pkg/core"
    "fmt"
    "os"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]
    fibonacciNum := os.Args[2]

    env := make(map[string]string)
    env["fibonacciNum"] = fibonacciNum

    processSpec := core.CreateProcessSpec(colonyID, []string{}, "FibonacciSolver", -1, 3, 1000, 10, 1, env)

    client := client.CreateColoniesClient("localhost", 8080, true)
    addedProcess, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(addedProcess.ID)
}
```

Let's generate some task to the solver.
```console
go run generator2.go a7c4bd22e94010027b31bf669c766aed0e24d6e4e1da05511b56440238df108f 43
```

Output (process Id):
```
99b3a2ad38acc451116604c6b98aa05bec1b7496ecd7d83a0f42fd049a6e04ab
```

Let's use CLI to look up the process.
```console
colonies process get --processid 99b3a2ad38acc451116604c6b98aa05bec1b7496ecd7d83a0f42fd049a6e04ab
```

Output:
```
Process:
+-------------------+------------------------------------------------------------------+
| ID                | 99b3a2ad38acc451116604c6b98aa05bec1b7496ecd7d83a0f42fd049a6e04ab |
| IsAssigned        | True                                                             |
| AssignedRuntimeID | 964cb72ef11deab2f66f089fdbd5d4b5df934af0774aa67fb0695e196df4d1a8 |
| Status            | Successful                                                       |
| SubmissionTime    | 2022-01-03T21:02:43.857048Z                                      |
| StartTime         | 2022-01-03T21:02:46.546153Z                                      |
| EndTime           | 2022-01-03T21:02:46.586151Z                                      |
| Deadline          | 0001-01-01T00:00:00Z                                             |
| Retries           | 0                                                                |
+-------------------+------------------------------------------------------------------+

Requirements:
+----------------+------------------------------------------------------------------+
| ColonyID       | 50d74cb4c8306856a4c854089280c6be80353b36e6f61c41c435f8c87c9ec1cb |
| RuntimeIDs     | None                                                             |
| RuntimeType    | FibonacciSolver                                                  |
| Memory         | 1000                                                             |
| CPU Cores      | 10                                                               |
| Number of GPUs | 1                                                                |
| Timeout        | -1                                                               |
| Max retries    | 3                                                                |
+----------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+--------------+-----------+------+
|                                ID                                |     KEY      |   VALUE   | TYPE |
+------------------------------------------------------------------+--------------+-----------+------+
| 5a8d3b7f8115a8812f1d7edb135862b61ba7a5b1828a26f5b7abff0d5acb15b3 | fibonacciNum |        43 | Env  |
| 02b4e846823547a7cb7e2f942c381929d757b7f8b91464e843db1c0c90d2091f | result       | 433494437 | Out  |
+------------------------------------------------------------------+--------------+-----------+------+
```

Note the **result** value.

## 11. Let the generator wait for the process to finish (generator_sub.go)

This can be done with this additional code:
```go
 subscription, _ := client.SubscribeProcess(addedProcess.ID, core.SUCCESS, 100, runtimePrvKey)
 process := <-subscription.ProcessChan
```

Note the **<-subscription.ProcessChan** will block until the process finishes.

```go
import (
    "colonies/pkg/client"
    "colonies/pkg/core"
    "fmt"
    "os"
)

func main() {
    colonyID := os.Getenv("COLONYID")
    runtimePrvKey := os.Args[1]
    fibonacciNum := os.Args[2]

    env := make(map[string]string)
    env["fibonacciNum"] = fibonacciNum

    processSpec := core.CreateProcessSpec(colonyID, []string{}, "FibonacciSolver", -1, 3, 1000, 10, 1, env)

    client := client.CreateColoniesClient("localhost", 8080, true)
    addedProcess, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
    if err != nil {
        fmt.Println(err)
    }

    subscription, _ := client.SubscribeProcess(addedProcess.ID, core.SUCCESS, 100, runtimePrvKey)
    process := <-subscription.ProcessChan

    for _, attribute := range process.Attributes {
        if attribute.Key == "result" {
            fmt.Println(attribute.Value)
        }
    }
}
```

In one terminal type:
```console
go run solver_sub.go dc99f66eb882813916135fc0f1d913c38b2ac0435f3e3fae60eb70b421a92e28
```

In another terminal type: 
```console
go run generator_ret.go a7c4bd22e94010027b31bf669c766aed0e24d6e4e1da05511b56440238df108f 43
```

After the generator has finished the first terminal will print **433494437**.
