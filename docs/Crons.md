# Using cron to spawn workflows
Generators are useful when there is a sequence of incoming data, and a workflow should be spawned when there is a certain amount of data. 
Cron on the other hand are spawned periodically by specifying a cron expression or an interval. It is also possible to randomly spawn workflows within a specified interval. 

## Cron expressions
Cron expressions follow this format:
```
┌───────────── second (0 - 59)
│ ┌───────────── minute (0 - 59) 
│ │ ┌───────────── hour (0 - 23)
│ │ │ ┌───────────── day of the month (1 - 31)
│ │ │ │ ┌───────────── month (1 - 12) 
│ │ │ │ │ ┌───────────── day of the week
│ │ │ │ │ │ 
│ │ │ │ │ │ 
* * * * * *
```

Spawn a workflow every second starting at 00 seconds: 
```
0/1 * * * * *
```

Spawn a workflow every other second starting at 00 seconds: 
```
0/2 * * * * *
```

Spawn a workflow every minute starting at 30 seconds: 
```
30 * * * * *
```

Spawn a workflow every Monday at 15:03:59: 
```
59 3 15 * * MON
```

Spawn a workflow every Christmas Eve at 15:00: 
```
0 0 15 24 12 * 
```

## Adding a cron  
Cron workflows can either be added using the Colonies API/SDK or by using the CLI:

```console
colonies cron add --name example_cron --cron "0/5 * * * * *" --spec examples/cron/cron_workflow.json 
```

Output:
```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
INFO[0000] Cron added                                    CronID=e2c81ec5b2ab75c2290cf195310105f1e8f5f1b733b70973f843dc2adb7708ac
```

The submitted workflow consists of two processes. The first process (generate_date) stores the current time to a file (/tmp/currentdate). The seconds process (print_date), which can not start before first process has finished, printed the /tmp/currentdate file to standard out.

```json
[
    {
        "nodename": "generate_date",
        "funcname": "date",
        "args": [
            ">",
            "/tmp/currentdate"
        ],
        "conditions": {
            "executortype": "cli",
            "dependencies": []
        }
    },
    {
        "nodename": "print_date",
        "funcname": "cat",
        "args": [
            "/tmp/currentdate"
        ],
        "conditions": {
            "executortype": "cli",
            "dependencies": [
                "generate_date"
            ]
        }
    }
]
```

### Spawn a new workflow on Christmas
```console
colonies cron add --name christmas_cron --cron "0 0 15 24 12 *" --spec examples/cron/christmas_workflow.json 
```

Output:
```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
INFO[0000] Cron added                                    CronID 5509ae48af5ecbd6d3b395ae2cd8c5bf0ad9ef7d83a5abb4328acea090c62b66
```

## Getting info about a cron 
```console
colonoies cron get --cronid e2c81ec5b2ab75c2290cf195310105f1e8f5f1b733b70973f843dc2adb7708ac
```

Output:
```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
Cron:
+----------------------+------------------------------------------------------------------+
| Id                   | e2c81ec5b2ab75c2290cf195310105f1e8f5f1b733b70973f843dc2adb7708ac |
| ColonyID             | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| Name                 | example_cron                                                     |
| Cron Expression      | 0/5 * * * * *                                                    |
| Interval             | -1                                                               |
| Random               | false                                                            |
| NextRun              | 2022-08-20 20:05:45                                              |
| LastRun              | 2022-08-20 20:05:40                                              |
| Last known WorflowID | 1e6cdf2fa1dd5602392c0e43d6f4473a60ae201d9476a3a7cc545d3c7ae022a5 |
+----------------------+------------------------------------------------------------------+

WorkflowSpec:

FunctionSpec 0:
+-------------+---------------------+
| FuncName    | date               |
| Args        | > /tmp/currentdate |
| MaxWaitTime | 0                  |
| MaxExecTime | 0                  |
| MaxRetries  | 0                  |
| Priority    | 0                  |
+-------------+---------------------+

Conditions:
+--------------+------+
| ColonyID     |      |
| ExecutorIDs  | None |
| ExecutorType | cli  |
| Dependencies |      |
+--------------+------+

FunctionSpec 1:
+-------------+-------------------+
| FuncName    | cat              |
| Args        | /tmp/currentdate |
| MaxWaitTime | 0                |
| MaxExecTime | 0                |
| MaxRetries  | 0                |
| Priority    | 0                |
+-------------+-------------------+

Conditions:
+--------------+---------------+
| ColonyID     |               |
| ExecutorIDs  | None          |
| ExecutorType | cli           |
| Dependencies | generate_date |
+--------------+---------------+
```

## Lists all available cron 
```console
colonies cron ls 
```

Output:
```console
+------------------------------------------------------------------+----------------+
|                              CRONID                              |      NAME      |
+------------------------------------------------------------------+----------------+
| 5509ae48af5ecbd6d3b395ae2cd8c5bf0ad9ef7d83a5abb4328acea090c62b66 | cristmast_cron |
| 4a5bbb434ff4c129ca53f22877ffe55a34fdea8758debdcce514fb6c40310ec4 | example_cron   |
+------------------------------------------------------------------+----------------+
```

## Immediately run a cron
Don't wait for Santa!

```console
colonies cron run --cronid 5509ae48af5ecbd6d3b395ae2cd8c5bf0ad9ef7d83a5abb4328acea090c62b66 
```

Output:
```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
INFO[0000] Running cron                                  CronID=
```

There are now quite a few processes in the queue:

```console
colonies process psw
```

Output:
```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
+------------------------------------------------------------------+------+--------------------+---------------------+---------------+
|                                ID                                | FUNC |        ARGS        |   SUBMISSION TIME   | EXECUTOR TYPE |
+------------------------------------------------------------------+------+--------------------+---------------------+---------------+
| b8a92460b765c66df7e3d81094c6f23ba0c93482699264b7bf1ca81e3d09c6ae | cat  | /tmp/currentdate   | 2022-08-20 20:34:35 | cli           |
| 91f852ca62bf302fdd7d9dc76632e2e230c3dac98af79c8ea05ea755106e44c1 | date | > /tmp/currentdate | 2022-08-20 20:34:35 | cli           |
| 8688d619854fd3a100ecb8527a2983905976577bde2017f772f34cec1fc42e44 | echo | HU HU HU           | 2022-08-20 20:34:33 | cli           |
| 4a853216f8f08a6cecc6b3bf8ac2190ce92cb51d7113f0b2a585c5eed0883c12 | cat  | /tmp/currentdate   | 2022-08-20 20:34:30 | cli           |
| bc5ddcf740cabab6608bc9e2a6cb02035d87bf101fb15bd1ec96d6b683308cfd | date | > /tmp/currentdate | 2022-08-20 20:34:30 | cli           |
| 67894f91a94b09e48cae77b87ce6cbfb0032ab982967b9acfdc4f5a7de6ddac7 | cat  | /tmp/currentdate   | 2022-08-20 20:34:25 | cli           |
| 529426e0fa32bc1cc42bc3b7cf1d93662cc75054aec997b6cdd201b5043fc2a0 | date | > /tmp/currentdate | 2022-08-20 20:34:25 | cli           |
| 2e0d66a5fb9bab78b2bc48c795ed348c688844a8a79c3afb4445afc68192f1cf | cat  | /tmp/currentdate   | 2022-08-20 20:34:20 | cli           |
| d4f2fd7d7a0b97b7aa6f604446b38d284cbfa95510a7a598307549bc4d1bf8e6 | date | > /tmp/currentdate | 2022-08-20 20:34:20 | cli           |
| 224eaabe55544f72abfd647a511cc6bfb4c45d7e8e0594432bab2b84e034fc2c | cat  | /tmp/currentdate   | 2022-08-20 20:34:15 | cli           |
| 01f52fe8809f77c063696717f902ad4558b8c886448c490bd88b989d25fc9cfe | date | > /tmp/currentdate | 2022-08-20 20:34:15 | cli           |
| 700ced47f81e8bf7d1f848072e1d0641b066c9934800ba749c08280fe3252bfd | cat  | /tmp/currentdate   | 2022-08-20 20:34:10 | cli           |
| b71731d6179bdd68e095c0f15de55dc8cecffd9d7cba4941fd87e7ca5a6eeffb | date | > /tmp/currentdate | 2022-08-20 20:34:10 | cli           |
+------------------------------------------------------------------+------+--------------------+---------------------+---------------+
```

The waiting queue will just keep on increasing if there are no executors executing the processes. However, we can set **MaxWaitingTime** on the process spec so that processes are automatically removed if not executed after a certain amount of time. In the example below, they will be removed after 3 seconds if not executed.

```json
[
    {
        "nodename": "generate_date",
        "funcname": "date",
        "args": [
            ">",
            "/tmp/currentdate"
        ],
        "conditions": {
            "executortype": "cli",
            "dependencies": []
        },
        "maxwaittime": 3
    },
    {
        "nodename": "print_date",
        "funcname": "cat",
        "args": [
            "/tmp/currentdate"
        ],
        "conditions": {
            "executortype": "cli",
            "dependencies": [
                "generate_date"
            ]
        },
        "maxwaittime": 5
    }
]
```

## Use interval instead of a cron expressions 
An alternative way to spawn a cron is to specify an interval instead of a cron expression. In the example, below a workflow is spawned every 10 seconds.

```console
 cron add --name example_cron --interval 10 --spec examples/cron/cron_maxwaittime_workflow.json
```

## Random intervals
It is also possible to spawn a workflow at a random time within an interval. This can be very useful when testing a software (e.g. chaos engineering).

```console
 cron add --name example_cron --interval 10 --random --spec examples/cron/cron_maxwaittime_workflow.json
```

In the example, a workflow will be spawned randomly within 10 seconds. Use the get command to find out the next time (NextRun) it will run.
```console
colonies cron get --cronid bb345cca6eb919824989a169f589b508841d6aaa4b020377da624afb2e7af9fe  
```

Output:
```console
+----------------------+------------------------------------------------------------------+
| Id                   | bb345cca6eb919824989a169f589b508841d6aaa4b020377da624afb2e7af9fe |
| ColonyID             | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| Name                 | example_cron                                                     |
| Cron Expression      |                                                                  |
| Interval             | 10                                                               |
| Random               | true                                                             |
| NextRun              | 2022-08-20 21:00:35                                              |
| LastRun              | 2022-08-20 21:00:25                                              |
| Last known WorflowID | 0aa1bb9ffe4e40b140689ad902012358b64c0bc48d3b6aa1fc5d5975a3530f70 |
+----------------------+------------------------------------------------------------------+
```

## Delete a cron
```console
colonies cron delete --cronid  ba6e938289b8e33c399678f9b812af0c3602a36704841965c2dc8c672efc1834 
```
