# Generators
Generators receive data (strings) from clients (via API) and automatically spawn workflows when number of calls exceed a threshold. Data is then available as an argument to the Colonies process.

## Start a Colonies server
```console
colonies dev
```
## Add a generator
```console
colonies generator add --spec ./examples/generator_workflow.json --name testgenerator --trigger 5
```
Output:
```
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
INFO[0000] Generator added                               GeneratorID=f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7
```

The workflow just echo the args. It look like this:
```json
[
    {
        "name": "task_a",
        "func": "echo",
        "args": [],
        "conditions": {
            "runtimetype": "cli",
            "dependencies": []
        }
    }
]
```

## Start a worker executing the workflows spawn by the generator, i.e. run the Unix echo command.
```console
colonies worker start --name generator_worker --runtimetype cli --timeout 100 -v 
```

## Send data to the Generator 
```console
colonies generator pack --generatorid f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7 --arg hello1
colonies generator pack --generatorid f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7 --arg hello2
colonies generator pack --generatorid f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7 --arg hello3
colonies generator pack --generatorid f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7 --arg hello4
colonies generator pack --generatorid f3a433d0a428ddd21fba2b82659db40dfc4e70771a29e2a19743ad80033749d7 --arg hello5
```

Notice that a workflow is spawn after the last pack call, as number of pack calls > trigger. In the worker terminal we can see:

```console
INFO[0312] Worker was assigned a process                 ProcessID=3806424831e78001fd7157a387ca9ab414ef908f0649eeed7e9fee691438db01
INFO[0312] Lauching process                              Args="[hello1 hello2 hello3 hello4 hello5]" Func=echo
hello1 hello2 hello3 hello4 hello5
INFO[0312] Closing process as successful                 processID=3806424831e78001fd7157a387ca9ab414ef908f0649eeed7e9fee691438db01
```

If we look up the process we get:
```console
colonies process get --processid 3806424831e78001fd7157a387ca9ab414ef908f0649eeed7e9fee691438db01
```

Output:
```console
Process:
+-------------------+------------------------------------------------------------------+
| ID                | 3806424831e78001fd7157a387ca9ab414ef908f0649eeed7e9fee691438db01 |
| IsAssigned        | True                                                             |
| AssignedRuntimeID | eeefa45d65b75c6ec3e11fedd2b120909607da830bade4f1953e55ccbad417c1 |
| State             | Successful                                                       |
| Priority          | 0                                                                |
| SubmissionTime    | 2022-08-23 22:19:14                                              |
| StartTime         | 2022-08-23 22:19:14                                              |
| EndTime           | 2022-08-23 22:19:14                                              |
| WaitDeadline      | 0001-01-01 01:12:12                                              |
| ExecDeadline      | 0001-01-01 01:12:12                                              |
| WaitingTime       | 9.712ms                                                          |
| ProcessingTime    | 12.305ms                                                         |
| Retries           | 0                                                                |
| ErrorMsg          |                                                                  |
+-------------------+------------------------------------------------------------------+

ProcessSpec:
+-------------+--------------------------------+
| Func        | echo                           |
| Args        | hello1 hello2 hello3 hello4    |
|             | hello5                         |
| MaxWaitTime | 0                              |
| MaxExecTime | -1                             |
| MaxRetries  | 0                              |
| Priority    | 0                              |
+-------------+--------------------------------+

Conditions:
+--------------+------------------------------------------------------------------+
| ColonyID     | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| RuntimeIDs   | None                                                             |
| RuntimeType  | cli                                                              |
| Dependencies |                                                                  |
+--------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+--------+--------------------------------+------+
|                                ID                                |  KEY   |             VALUE              | TYPE |
+------------------------------------------------------------------+--------+--------------------------------+------+
| b4159b8813a657d20b88ad7231bc388cbc2b1e296e1c5bf02e28c44486ac95b1 | output | hello1 hello2 hello3 hello4    | Out  |
|                                                                  |        | hel...                         |      |
+------------------------------------------------------------------+--------+--------------------------------+------+
```
