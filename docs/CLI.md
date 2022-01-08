# Using the Colonies CLI tool 
## Register a new Colony
First, create a file named colony.json, and put the following content into it.
```json
{
    "name": "mycolony"
}
```

Then use the colonies tool to register the colony. The id of the colony will be returned if the command is successful. Note that the root password is required for this operation.
```console
./bin/colonies colony register --serverid=9289dfccedf27392810b96968535530bb69f90afe7c35738e0e627f3810d943e --spec ./examples/colony.json
```
Output: 
```
0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5
```

## List all Colonies 
Note that root password of Colonies server is also required to list all colonies.
```console
./bin/colonies colony ls --serverid=9289dfccedf27392810b96968535530bb69f90afe7c35738e0e627f3810d943e 
```
Output:
```
+------------------------------------------------------------------+----------+
|                                ID                                |   NAME   |
+------------------------------------------------------------------+----------+
| 0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5 | mycolony |
+------------------------------------------------------------------+----------+
```

## Register a new Colony Runtime 
Only the colony owner is allowed to register a new Colony Runtime. 

```json
{
    "name": "my_runtime",
    "runtimetype": "my_runtime_type",
    "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
    "cores": 32,
    "mem": 80326,
    "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
    "gpus": 1
}
```

```console
./bin/colonies runtime register --colonyid 0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5 --colonyprvkey d95c54b63ac7c9ba624445fd755998e14e6aa71a17a74889c6a1754be80bcf09 --spec ./examples/runtime.json
```
Output:
```
The *colonyprvkey* is automatically obtained from the keychain if not specified. The *colonyid* can also be specified using an environmental variables.
```

```console
export COLONYID="0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5"
./bin/colonies runtime register --spec ./examples/runtime.json
```
Output:
```
4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58
```

## List registered Colony Runtimes
```console
export RUNTIMEID="4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58"
./bin/colonies runtime ls 
```
Output:
```
Runtime with Id <4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58> is not approved
```

A Colony Runtime needs to be approved by the Colony Owner before it can execute processes. As before, the Colony Owner's private key is automatically fetched from the keychain.

```console
./bin/colonies runtime approve --runtimeid 4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58 
```
Output:
```
Colony Runtime with Id <4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58> is now approved
```

```console
./bin/colonies runtimes ls 
```
Output:
```
+------------------------------------------------------------------+------------+----------+
|                                ID                                |    NAME    |  STATE   |
+------------------------------------------------------------------+------------+----------+
| 4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58 | my_runtime | Approved |
+------------------------------------------------------------------+------------+----------+
```

Similarly, a Colony Runtime can be rejected with the "rejected" command. 
```console
./bin/colonies runtime reject --runtimeid 4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58 
```
Output:
```
Colony Runtime with Id <4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58> is now rejected
```

## Submit a process to a Colony
First we need to create a process spec file. The conditions must match registered runtimes, e.g. the memory must be at least 1000 GiB.

```json
{
     "conditions": {
         "colonyid": "0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5",
         "runtimeids": [],
         "runtimetype": "my_runtime_type",
         "mem": 1000,
         "cores": 10,
         "gpus": 1
     },
     "env": {
         "test_key": "test_value"
     }
     "timeout": -1,
     "maxretries": 3
}
```

To submit the process spec to the Colony, type:

```console
./bin/colonies process submit --spec ./examples/process_spec.json
```
Output:
```
7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2
```

## Get info about a process
```console
./bin/colonies process get --processid 4e369a9eeaf4521cdfa79de81666a5980f30345464e5c61e8cfdf9380e7ba663 
```
Output:
```
Process:
+-------------------+------------------------------------------------------------------+
| ID                | 4e369a9eeaf4521cdfa79de81666a5980f30345464e5c61e8cfdf9380e7ba663 |
| IsAssigned        | True                                                             |
| AssignedRuntimeID | 4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58 |
| State             | Running                                                          |
| SubmissionTime    | 2021-12-28T16:26:33.838548Z                                      |
| StartTime         | 2021-12-28T17:05:12.228424Z                                      |
| EndTime           | 0001-01-01T00:00:00Z                                             |
| Deadline          | 0001-01-01T00:00:00Z                                             |
| Retries           | 0                                                                |
+-------------------+------------------------------------------------------------------+

Requirements:
+----------------+------------------------------------------------------------------+
| ColonyID       | 0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5 |
| RuntimeIDs     | None                                                             |
| RuntimeType    | my_runtime_type                                                  |
| Memory         | 1000                                                             |
| CPU Cores      | 10                                                               |
| Number of GPUs | 1                                                                |
| Timeout        | -1                                                               |
| Max retries    | 3                                                                |
+----------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+----------+------------+------+
|                                ID                                |   KEY    |   VALUE    | TYPE |
+------------------------------------------------------------------+----------+------------+------+
| 7127634e101022509a16951658677a6a7f10a9b045e8cd4491501b5c6888ee3a | test_key | test_value | Env  |
+------------------------------------------------------------------+----------+------------+------+
```

## List all waiting processes
```console
./bin/colonies process psw
```
Output:
```
+------------------------------------------------------------------+-----------------------------+
|                                ID                                |       SUBMISSION TIME       |
+------------------------------------------------------------------+-----------------------------+
| 25c3fbf4c7ad4558458f86db61353988e2a88803014b530c3952f7f9fbbf2244 | 2021-12-28T15:31:08.369233Z |
| 5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258 | 2021-12-28T15:34:46.548835Z |
| aca88cd66d96a1acce0897f9485abc4d072ab52bed388bdbddf4ebffaf862f84 | 2021-12-28T15:37:12.813552Z |
| 4e369a9eeaf4521cdfa79de81666a5980f30345464e5c61e8cfdf9380e7ba663 | 2021-12-28T16:26:33.838548Z |
| 7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2 | 2021-12-28T16:29:44.153707Z |
+------------------------------------------------------------------+-----------------------------+
```

## Assign a process to runtime 
An assigned process will change state to Running.
```console
./bin/colonies process assign
```
Output:
```
Process with Id <5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258> was assigned to Runtime with Id <4599f89a8afb7ecd9beec0b7861fab3bacba3a0e2dbe050e9f7584f3c9d7ac58>
```

## List all running processes
```console
./bin/colonies process ps
```
Output:
```
+------------------------------------------------------------------+-----------------------------+
|                                ID                                |         START TIME          |
+------------------------------------------------------------------+-----------------------------+
| 25c3fbf4c7ad4558458f86db61353988e2a88803014b530c3952f7f9fbbf2244 | 2021-12-28T17:01:31.363053Z |
| 5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258 | 2021-12-28T17:03:51.557583Z |
| aca88cd66d96a1acce0897f9485abc4d072ab52bed388bdbddf4ebffaf862f84 | 2021-12-28T17:05:11.723638Z |
| 4e369a9eeaf4521cdfa79de81666a5980f30345464e5c61e8cfdf9380e7ba663 | 2021-12-28T17:05:12.228424Z |
| 7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2 | 2021-12-28T17:05:12.547542Z |
+------------------------------------------------------------------+-----------------------------+
```

## List all successful processes
```console
./bin/colonies process pss 
```
Output:
```
+------------------------------------------------------------------+-----------------------------+
|                                ID                                |          END TIME           |
+------------------------------------------------------------------+-----------------------------+
| 25c3fbf4c7ad4558458f86db61353988e2a88803014b530c3952f7f9fbbf2244 | 2021-12-28T17:22:46.17229Z  |
| 5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258 | 2021-12-28T17:24:01.479675Z |
+------------------------------------------------------------------+-----------------------------+
```

## List all failed processes
```console
./bin/colonies process psf 
```
Output:
```
+------------------------------------------------------------------+-----------------------------+
|                                ID                                |          END TIME           |
+------------------------------------------------------------------+-----------------------------+
| 7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2 | 2021-12-28T17:25:05.112377Z |
+------------------------------------------------------------------+-----------------------------+
```

## Add attribute to a running process 
```console
./bin/colonies attribute add --key output --value helloworld --processid 5785eb8a57f22d73a99d5c5e5d073cf27f9ea4ba81bad1a72e5e4f226e647dc0 
```

Output:
```
7fcc3a10947e6a3c56fa5c59c14c7d13d32468ed899e12e9d1cb7589ef51a0e3
```

```console
./bin/colonies process get --processid 5785eb8a57f22d73a99d5c5e5d073cf27f9ea4ba81bad1a72e5e4f226e647dc0
```
Output:
```
+-------------------+------------------------------------------------------------------+
| ID                | 5785eb8a57f22d73a99d5c5e5d073cf27f9ea4ba81bad1a72e5e4f226e647dc0 |
| IsAssigned        | False                                                            |
| AssignedRuntimeID | None                                                             |
| State             | Waiting                                                          |
| SubmissionTime    | 2021-12-28T17:40:45.749629Z                                      |
| StartTime         | 0001-01-01T00:00:00Z                                             |
| EndTime           | 0001-01-01T00:00:00Z                                             |
| Deadline          | 0001-01-01T00:00:00Z                                             |
| Retries           | 0                                                                |
+-------------------+------------------------------------------------------------------+

Requirements:
+----------------+------------------------------------------------------------------+
| ColonyID       | 0f4f350d264d1cffdec0d62c723a7da8b730c6863365da75697fd26a6d79ccc5 |
| RuntimeIDs     | None                                                             |
| RuntimeType    | my_runtime_type                                                  |
| Memory         | 1000                                                             |
| CPU Cores      | 10                                                               |
| Number of GPUs | 1                                                                |
| Timeout        | -1                                                               |
| Max retries    | 3                                                                |
+----------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+------------+-------------+------+
|                                ID                                |    KEY     |    VALUE    | TYPE |
+------------------------------------------------------------------+------------+-------------+------+
| 1193364beaddb9e33557776fe3d2574459e0e0ca73733422d00fc0f1e4f6ccb2 | test_key   | test_value  | Env  |
| 7fcc3a10947e6a3c56fa5c59c14c7d13d32468ed899e12e9d1cb7589ef51a0e3 | output     | hello world | Out  |
+------------------------------------------------------------------+------------+-------------+------+
```

## Get attribute of a process 
```console
./bin/colonies attribute get --attributeid 7fcc3a10947e6a3c56fa5c59c14c7d13d32468ed899e12e9d1cb7589ef51a0e3 --processid 5785eb8a57f22d73a99d5c5e5d073cf27f9ea4ba81bad1a72e5e4f226e647dc0 
```
Output:
```
+---------------+------------------------------------------------------------------+
| ID            | 7fcc3a10947e6a3c56fa5c59c14c7d13d32468ed899e12e9d1cb7589ef51a0e3 |
| TargetID      | 5785eb8a57f22d73a99d5c5e5d073cf27f9ea4ba81bad1a72e5e4f226e647dc0 |
| AttributeType | Out                                                              |
| Key           | output                                                           |
| Value         | hello world                                                      |
+---------------+------------------------------------------------------------------+
```

## Close a process as successful
```console
./bin/colonies process successful --processid 5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258
```
Output:
```
Process with Id <5513617dc4407b6190959a07db2a39c6ad93771c7e8457391e2e64927214c258> closed as successful
```

## Close a process as failed 
```console
./bin/colonies process failed --processid 7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2
```
Output:
```
Process with Id <7bdc97997db5ea59471b2165c0e5672a4fe8f9158d36ab547adb9710d26e5ae2> closed as failed
```
