[![codecov](https://codecov.io/gh/colonyos/colonies/branch/main/graph/badge.svg?token=1D4O2JVSJL)](https://codecov.io/gh/colonyos/colonies)
[![Go](https://github.com/colonyos/colonies/actions/workflows/go.yml/badge.svg)](https://github.com/colonyos/colonies/actions/workflows/go.yml)

![ColonyOSLogo](docs/images/ColonyOsLogoNoShaddow2.png)

# What is Colonies?
To simplify, Colonies is an **Employment Agency** for Internet-connected computers. Humans (or machines) submit job specifications to a Colonies server. The Colonies server keeps track of one or several Colonies, which are communities of machines. Machines (co-called workers) connect to the Colonies server and search for suitable jobs in Colonies they are member of. Each worker must have a valid identity to prove its membership and the Colonies server makes sure only authorized and qualified workers can connect and be assigned relevant jobs.

* A Colonies worker can **reside anywhere on the Internet**, e.g. inside a Kubernetes Pod, a smart phone, or embedded in a web page, enabling a compute continuum spannig devices, edge and cloud.
* Colonies makes it possible to **orchestrate processes inside containers**. This is far more efficient than lauching a new container for each new job.
* Colonies is **fast**. Instead of using RAFT for distributed consensus, Colonies uses a standard PostgreSQL database for storing states and execution history.   
* A Colonies server is stateless and designed to take **full advantage of Kubernetes**, e.g. fault tolerance and scalability. 
* Colonies **complements Kubernetes** and brings robust and fault tolerant **batch processing** to Kubernetes, typically needed by many AI workloads.
* Colonies is **lightweight and works without Kubernetes**. It runs in browsers, smart phones or IoT devices. This also makes it much easier to test and develop complex workflows before they are deployed on Kubernetes.
* Colonies **builds on security**, a  built-in crypto-protocol ensure secure and **zero-trust process execution**. 
* Colonies is currently being used by **[RockSigma AB](https://www.rocksigma.com)** to build a novel compute engine for automatic seismic processing in underground mines. 

![Colonies Architecture](docs/images/ColoniesArchFull.png?raw=true "Colonies Architecture")

More information can also be found [here](https://colonyos.io).

## Key features
* **Batch processing and distributed RPC.** The Colonies server maintains several prioritized job queues and keeps track of process statuses. Processes not finishing in time are automatically moved back to the job queue to be executed by another worker.  
* **Pull-based orchestration.** Users (or workers) submit process specifications the Colonies server. Colonies workers connect to the Colonies server and request processes to execute. A HTTP Long Polling/WebSocket protocol ensure that workers can reside anywhere on the Internet, even behind firewalls. The Colonies server never establish connections directly to workers. 
* **Multi-step workflows** or **Directed Acyclic Graph (DAG)** to capture dependencies between jobs.
* **Built-in identity and trust management.** A crypto-protocol based on ECDSA (Elliptic Curve Digital Signature Algorithm) offers identity and trust management to enable Colonies workers member of the same colony to fully trust each other. Only authorized users or workers can submit process specifications or interact with other workers within a colony.
* **Implemented in Golang** with a standard PostgreSQL database.
* **SDK in Python, Julia, and Golang.**

## Example of use cases
* **Manage ML/AI workloads on Kubernetes.** Launch one or several Colonies worker containers in a Kubernetes Pod. Then use Colonies to enable batch processing and launch processes inside worker containers. Launching processes inside already started containers can be significantly more efficient than frameworks like [Argo Workflows](https://argoproj.github.io/argo-workflows) that launches new containers for each new job, especially when dealing with AI workflows consisting of huge containers (tens of gigabytes) or when a huge amount of data needs to be shuffled into memory to perform a certain computation.
* **Grid computing.** Create "non-malicious" botnets and launch processes to perform computations at IoT devices, smart phones or cloud servers; all controlled from the Colonies server.
* **Manage complex workflows spanning multiple cloud/edge servers and devices**, e.g. setting up multimedia pipelines and ML inference servers running on multiple platforms connected to different networks.

# Links
* [Installation](docs/Installation.md)
* [Using the Colonies CLI tool](docs/CLI.md)
* [Tutorial (implement your own Colonies Worker using the Golang SDK)](docs/Tutorial.md)
* [HTTP RPC Protocol](docs/RPC.md)

# Getting started example
## Install Colonies CLI tool
```console
wget https://github.com/colonyos/colonies/blob/main/bin/colonies?raw=true -O /bin/colonies
chmod +x /bin/colonies
```

## Start a Colonies development server
The development server is for testing only. All data will be lost when restarting the server. Also note that all keys are well known and all data is sent over unencrypted HTTP.

You will have to export some environmental variables in order to use the server. Add these variables to your shell.

```console
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8
export LC_CTYPE=UTF-8
export TZ=Europe/Stockholm
export COLONIES_TLS="false"
export COLONIES_SERVERHOST="localhost"
export COLONIES_SERVERPORT="50080"
export COLONIES_SERVERID="039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103"
export COLONIES_SERVERPRVKEY="fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d"
export COLONIES_DBHOST="localhost"
export COLONIES_DBUSER="postgres"
export COLONIES_DBPORT="50070"
expoer COLONIES_DBPASSWORD="rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
export COLONIES_COLONYID="4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
export COLONIES_COLONYPRVKEY="ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514"
export COLONIES_RUNTIMEID="3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac"
export COLONIES_RUNTIMEPRVKEY="ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"
export COLONIES_RUNTIMETYPE="cli"
```
or 
```console
source examples/devenv
```

Now, start the development server. The development server will automatically add the keys from the environment (e.g. COLONIES_RUNTIMEPRVKEY) to the Colonies keychain.
```console
colonies dev
```

## Start a Colonies worker
Open another terminal (and *source examples/devenv*).

```console
colonies worker start --name myworker --type testworker
```
## Submit a process specification
Example process specification (see examples/sleep.json). The Colonies worker will pull the process specification from the Colonies dev server and start a *sleep* process. This will cause the worker above to sleep for 100s. The *env* array in the JSON below will automatically be exported as real environment variables in the sleep process.
```json
{
  "conditions": {
    "runtimetype": "testworker"
  },
  "cmd": "sleep",
  "args": [
    "100"
  ],
  "env": {
    "TEST": "testenv"
  }
}
```

Open another terminal (and *source examples/devenv*).
```console
colonies process submit --spec sleep.json
```

Check out running processes:
```console
colonies process ps
+------------------------------------------------------------------+-------+------+---------------------+----------------+
|                                ID                                |  CMD  | ARGS |     START TIME      | TARGET RUNTIME |
+------------------------------------------------------------------+-------+------+---------------------+----------------+
| 6681946db095e0dc2e0408b87e119c0d2ae4f691db6899b829161fc97f14a1d0 | sleep | 100  | 2022-04-05 16:40:01 | testworker     |
+------------------------------------------------------------------+-------+------+---------------------+----------------+
```

Check out process status: 
```console
colonies process get --processid 6681946db095e0dc2e0408b87e119c0d2ae4f691db6899b829161fc97f14a1d0
Process:
+-------------------+------------------------------------------------------------------+
| ID                | 6681946db095e0dc2e0408b87e119c0d2ae4f691db6899b829161fc97f14a1d0 |
| IsAssigned        | True                                                             |
| AssignedRuntimeID | 66f55dcb577ca6ed466ad5fcab868673bc1fc7d6ea7db71a0af4fea86035c431 |
| State             | Running                                                          |
| SubmissionTime    | 2022-04-05 16:40:00                                              |
| StartTime         | 2022-04-05 16:40:01                                              |
| EndTime           | 0001-01-01 01:12:12                                              |
| Deadline          | 0001-01-01 01:12:12                                              |
| WaitingTime       | 753.441ms                                                        |
| ProcessingTime    | 1m23.585764152s                                                  |
| Retries           | 0                                                                |
+-------------------+------------------------------------------------------------------+

ProcessSpec:
+-------------+-------+
| Cmd         | sleep |
| Args        | 100   |
| Volumes     | None  |
| MaxExecTime | -1    |
| MaxRetries  | 0     |
+-------------+-------+

Conditions:
+-------------+------------------------------------------------------------------+
| ColonyID    | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| RuntimeIDs  | None                                                             |
| RuntimeType | testworker                                                       |
| Memory      | 0                                                                |
| CPU Cores   | 0                                                                |
| GPUs        | 0                                                                |
+-------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+------+---------+------+
|                                ID                                | KEY  |  VALUE  | TYPE |
+------------------------------------------------------------------+------+---------+------+
| 2fe15f1b570c7328854f2374a69e45845ef5a40624ec06c287a51a5732485ecc | TEST | testenv | Env  |
+------------------------------------------------------------------+------+---------+------+
```

## Execution time constraint
The *maxecution* attribute specifies the maxiumum execution time in seconds before the process specification (job) is moved back to the queue. The *maxretries* attributes specifies how many times it may be moved back to the queue. Execution time constraint is an import feature of Colonies to implement robust workflows. If a worker crash, the job will automatically moved back to the queue and be executed by another worker. 

This mechanism thus offer a last line of defense against failures and enables advanched software engineering disciplines such as [Chaos Engineering](https://en.wikipedia.org/wiki/Chaos_engineering). For example, a Chaos monkey may randomly kill worker pods in Kubernetes and Colonies guarantees that all jobs are eventually executed. 

```json
{
  "conditions": {
    "runtimetype": "testworker"
  },
  "cmd": "sleep",
  "args": [
    "100"
  ],
  "maxexectime": 5,
  "maxretries": 0,
  "env": {
    "TEST": "testenv"
  }
}
```

The process specification above will always result in failed Colonies processes as the the *sleep* process runs for exactly 100 seconds, but the process has to finish within 5 seconds. The *colonies process psf* command can be used to list all failed processes. 

```console
colonies process pss
WARN[0000] No successful processes found

colonies process psf
+------------------------------------------------------------------+-------+------+---------------------+--------------+
|                                ID                                |  CMD  | ARGS |      END TIME       | RUNTIME TYPE |
+------------------------------------------------------------------+-------+------+---------------------+--------------+
| 61789512c006fc132534d73d2ce5fd4a162f9b849548fcfe300bc5b8defa6400 | sleep | 100  | 2022-05-26 17:06:24 | testworker   |
+------------------------------------------------------------------+-------+------+---------------------+--------------+
```

Note that setting *maxretries* to -1 instead if 0 will result in a loop, a job that can never succeed.

## Workflow
A workflow is collection of named process specifications where some specifications may have dependencies to other specifications. Once submitted to the Colonies server, the Colonies server will create the corresponding processes and add the processes to the database (queue). It will also set dependencies between the processes which will then form a Directed Acyclic Graph (DAG). 

![ProcessGraph](docs/images/ProcessGraph.png)

The picture above depicts an example of a DAG. Task A has no depenecies and can thus start immediately. Task B and C have to wait for Task A to finish before they can start. Task D has to wait for Task B and C to finish. 

This workflow can be modelled as follows:
```json
{
    "processspecs": [{
            "name": "task_a",
            "cmd": "echo",
            "args": [
                "task1"
            ],
            "conditions": {
                "runtimetype": "cli",
                "dependencies": null
            }
        },
        {
            "name": "task_b",
            "cmd": "echo",
            "args": [
                "task2"
            ],
            "conditions": {
                "runtimetype": "cli",
                "dependencies": ["task_a"]
            }
        },
        {
            "name": "task_c",
            "cmd": "echo",
            "args": [
                "task3"
            ],
            "conditions": {
                "runtimetype": "cli",
                "dependencies": ["task_a"]
            }
        },
        {
            "name": "task_d",
            "cmd": "echo",
            "args": [
                "task4"
            ],
            "conditions": {
                "runtimetype": "cli",
                "dependencies": ["task_b", "task_c"]
            }
        }
    ]
}
```

## Submit a workflow 
Open another terminal (and *source examples/devenv*).
```console
colonies workflow submit --spec examples/workflow.json

INFO[0000] Workflow submitted                            WorkflowID=8bc49205ae35e089b370c05cd2a110b84e72d5052c2ec3fb5bc4832274d9d1b1
```

```console
colonies workflow get --workflowid 8bc49205ae35e089b370c05cd2a110b84e72d5052c2ec3fb5bc4832274d9d1b1

Workflow:
+----------------+------------------------------------------------------------------+
| WorkflowID     | 8bc49205ae35e089b370c05cd2a110b84e72d5052c2ec3fb5bc4832274d9d1b1 |
| ColonyID       | 8bc49205ae35e089b370c05cd2a110b84e72d5052c2ec3fb5bc4832274d9d1b1 |
| State          | Waiting                                                          |
| SubmissionTime | 2022-05-31 16:23:13                                              |
| StartTime      | 0001-01-01 01:12:12                                              |
| EndTime        | 0001-01-01 01:12:12                                              |
+----------------+------------------------------------------------------------------+

Processes:
+-------------------+------------------------------------------------------------------+
| Name              | task_a                                                           |
| ProcessID         | 3a8e9299c76905c87f903b4fdcf4c5dbeb314659e2ed31d477dcb414e8fedf1f |
| RuntimeType       | cli                                                              |
| Cmd               | echo                                                             |
| Args              | task_a                                                           |
| State             | Waiting                                                          |
| WaitingForParents | false                                                            |
| Dependencies      | None                                                             |
+-------------------+------------------------------------------------------------------+

+-------------------+------------------------------------------------------------------+
| Name              | task_b                                                           |
| ProcessID         | 5fd0611d57fc567ce7aa7984424b1de749c32b20b92668b4755ade6ca62e19c2 |
| RuntimeType       | cli                                                              |
| Cmd               | echo                                                             |
| Args              | task_b                                                           |
| State             | Waiting                                                          |
| WaitingForParents | true                                                             |
| Dependencies      | task_a                                                           |
+-------------------+------------------------------------------------------------------+

+-------------------+------------------------------------------------------------------+
| Name              | task_d                                                           |
| ProcessID         | f46b7e84da0657cda3982282f5bef8b3c7429eff6b635cbce9bf93eb034e6705 |
| RuntimeType       | cli                                                              |
| Cmd               | echo                                                             |
| Args              | task_d                                                           |
| State             | Waiting                                                          |
| WaitingForParents | true                                                             |
| Dependencies      | task_b task_c                                                    |
+-------------------+------------------------------------------------------------------+

+-------------------+------------------------------------------------------------------+
| Name              | task_c                                                           |
| ProcessID         | bf5d93190967539133063d357bcd5d446d3e4fce41a6d110926de12129a64156 |
| RuntimeType       | cli                                                              |
| Cmd               | echo                                                             |
| Args              | task_c                                                           |
| State             | Waiting                                                          |
| WaitingForParents | true                                                             |
| Dependencies      | task_a                                                           |
+-------------------+------------------------------------------------------------------+
```

## Start a worker
```console
colonies worker start --name myworker --type cli 

INFO[0000] Starting a worker                             BuildTime="2022-05-31T13:43:22Z" BuildVersion=a153cbf
INFO[0000] Saving runtimeID to /tmp/runtimeid
INFO[0000] Saving runtimePrvKey to /tmp/runtimeprvkey
INFO[0000] Register a new Runtime                        CPU= Cores=-1 GPU= GPUs=-1 Mem=-1 colonyID=4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 runtimeID=d709c23a58cb883817e0fe38ae20f3f539b7b7c4f607cc16e2b927eb3c123a34 runtimeName=myworker runtimeType:=cli
INFO[0000] Approving Runtime                             runtimeID=d709c23a58cb883817e0fe38ae20f3f539b7b7c4f607cc16e2b927eb3c123a34
INFO[0000] Worker now waiting for processes to be execute  BuildTime="2022-05-31T13:43:22Z" BuildVersion=a153cbf ServerHost=localhost ServerPort=50080
INFO[0000] Worker was assigned a process                 processID=3a8e9299c76905c87f903b4fdcf4c5dbeb314659e2ed31d477dcb414e8fedf1f
INFO[0000] Lauching process                              Args="[task_a]" Cmd=echo
task_a
INFO[0000] Closing process as successful                 processID=3a8e9299c76905c87f903b4fdcf4c5dbeb314659e2ed31d477dcb414e8fedf1f
INFO[0000] Worker was assigned a process                 processID=5fd0611d57fc567ce7aa7984424b1de749c32b20b92668b4755ade6ca62e19c2
INFO[0000] Lauching process                              Args="[task_b]" Cmd=echo
task_b
INFO[0000] Closing process as successful                 processID=5fd0611d57fc567ce7aa7984424b1de749c32b20b92668b4755ade6ca62e19c2
INFO[0000] Worker was assigned a process                 processID=bf5d93190967539133063d357bcd5d446d3e4fce41a6d110926de12129a64156
INFO[0000] Lauching process                              Args="[task_c]" Cmd=echo
task_c
INFO[0000] Closing process as successful                 processID=bf5d93190967539133063d357bcd5d446d3e4fce41a6d110926de12129a64156
INFO[0000] Worker was assigned a process                 processID=f46b7e84da0657cda3982282f5bef8b3c7429eff6b635cbce9bf93eb034e6705
INFO[0000] Lauching process                              Args="[task_d]" Cmd=echo
task_d
INFO[0000] Closing process as successful                 processID=f46b7e84da0657cda3982282f5bef8b3c7429eff6b635cbce9bf93eb034e6705
```

Note that the order the processes are executed. Also, try to start another worker and you will see that both workers will execute processes.

# Security principles
A core component of Colonies is a crypto identity protocol inspired by Bitcoin and Ethereum. Each colony and colony runtime is assigned a *Digital Identity* that is verified by the Colonies server using a so-called [Implicit certificates](https://en.wikipedia.org/wiki/Implicit_certificate), which is implemented using [Elliptic-curve cryptography](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography). This protocol makes it possible to reconstruct public-keys from signatures. Identities can then simply be calculated as cryptographic hashes (SHA3-256) of the reconstructed public-keys.

The Colonies server keeps track of these identities and applies several rules how runtimes are allowed to interact with each other. 

1. Only the server owner can register a new colony. 
2. Only the server owner can list registered colonies. 
3. Only a colony owner can register a colony runtimes to a colony. 
4. Only a colony owner can list/get info about colony.
5. Only a colony owner can approve/disapprove a runtime.
6. Any runtime member of a colony can submit/get/assign/list processes or workflows. 
7. Only the runtime that was assigned a process can set attributes on that process and close it. 
8. Any runtime can get/list attributes on processes. 

Note that the Colonies server does not store any crypto keys, but rather stores identites in a database and verifies that reconstructed identities obtained from RPC calls match the identities stored in the database. This protocol works as follows. Let's assume a runtime client has the following Id: 

```
69383f17554afbf81594999eec96adbaa0fc6caace5f07990248b14167c41e8f
```

It then sends the following message to the Colonies server:

```json
{
    "payloadtype": "addcolonymsg",
    "payload": "ewogICAgICBjb2xvbnlpZDogYWM4ZGM4OTQ5YWYzOTVmZDUxZWFkMzFkNTk4YjI1MmJkYTAyZjFmNmVlZDExYWNlN2ZjN2RjOGRkODVhYzMyZSwKICAgICAgbmFtZTogdGVzdF9jb2xvbnlfbmFtZQogIH0=",
    "signature": "82f2ba6368d5c7d0e9bfa6a01a8fa4d4263113f9eedf235e3a4c7b1febcdc2914fe1f8727746b2f501ceec5736457f218fe3b1a469dd6071775c472a802aa81501",
}
```
When the server receives the message, it reconstructs the Id of the calling client using the enclosed signature and payload. This means that client Id (e.g. 82f2ba6368d5c7d0e9bfa6...) is never sent to the server but rather derived by the server from messages it receives. In the example above, the server checks in the database if the reconstructed Id is a server owner.

# Running the Tests
Follow the instructions at [Installation Guide](./docs/Installation.md) and setup a Postgresql server, then type:
```console
make test
```

# Roadmap 1.0 release
* Cronjobs/job generators
* Built-in Distributed Hash table (Kademlia) to make it possible to create federations of Colonies servers to enable self-sovereignty and trust chains. Each Colonies server can join a global overlay network of connected Colonies servers on the Internet, similar to technologies like IPFS. 
