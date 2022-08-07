[![codecov](https://codecov.io/gh/colonyos/colonies/branch/main/graph/badge.svg?token=1D4O2JVSJL)](https://codecov.io/gh/colonyos/colonies)
[![Go](https://github.com/colonyos/colonies/actions/workflows/go.yml/badge.svg)](https://github.com/colonyos/colonies/actions/workflows/go.yml)

![ColonyOSLogo](docs/images/ColonyOsLogoNoShaddow2.png)

# What is Colonies?
**A Colony is like a Bee Colony, but for computer software**. 

A Colony is a trusted community of remotely connected computer programs (so-called workers) organized as a single unit to perform execution of various tasks. It is a platform for **process automation** and **distributed intelligence** and provides a **zero-trust infrastructure** for worker communication and task execution coordination within a Colony. The long-term vision is to create a global peer-to-peer network connecting many independent **self-sovereign** Colonies across the Internet. 

Simple helloworld worker in JavaScript:
```javascript
let colonyid = "4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
let runtimeid = "3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac"
let prvkey = "ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"

runtime.assign(colonyid, prvkey)
.then((process) => {
    if process.spec.func == "helloworld" {
        let attr = {
            targetid: process.processid,
            targetcolonyid: colonyid,
            attributetype: output,
            key: "result",
            value: "helloworld"
        } 
        runtime.addAttribute(attr, prvkey)
        runtime.closeProcess(process.processid, true, prvkey)
    }
})
.catch((err) => {
})
```

Submit a process using Colonies CLI:
```console
$ colonies process run --func helloworld --targettype cli
INFO[0000] Process submitted                             ProcessID=082c734838674d771452eb979d6fd3d05367ecf0cf4046a2254a56d5f32b3ef7
```

Lookup the result:
```console
$ colonies process get --processid 082c734838674d771452eb979d6fd3d05367ecf0cf4046a2254a56d5f32b3ef7
Process:
+-------------------+------------------------------------------------------------------+
| ID                | 082c734838674d771452eb979d6fd3d05367ecf0cf4046a2254a56d5f32b3ef7 |
| IsAssigned        | True                                                             |
| AssignedRuntimeID | 3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac |
| State             | Successful                                                       |
| Priority          | 0                                                                |
| SubmissionTime    | 2022-08-07 10:27:24                                              |
| StartTime         | 2022-08-07 10:27:24                                              |
| EndTime           | 2022-08-07 10:27:24                                              |
| WaitDeadline      | 2022-08-07 10:28:24                                              |
| ExecDeadline      | 0001-01-01 01:12:12                                              |
| WaitingTime       | 14.258ms                                                         |
| ProcessingTime    | 11.502ms                                                         |
| Retries           | 0                                                                |
| ErrorMsg          |                                                                  |
+-------------------+------------------------------------------------------------------+

ProcessSpec:
+-------------+--------------+
| Func        | setLampState |
| Args        | on           |
| MaxExecTime | -1           |
| MaxRetries  | -1           |
| Priority    | 0            |
+-------------+--------------+

Conditions:
+-------------+------------------------------------------------------------------+
| ColonyID    | 4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4 |
| RuntimeIDs  | None                                                             |
| RuntimeType | cli                                                              |
+-------------+------------------------------------------------------------------+

Attributes:
+------------------------------------------------------------------+--------+------------+------+
|                                ID                                |  KEY   |   VALUE    | TYPE |
+------------------------------------------------------------------+--------+------------+------+
| c6c1563a3fd6db649b820bad180b253eb317c6b0e9b4408580b4d8c9b948c33d | result | helloworld | Out  |
+------------------------------------------------------------------+--------+------------+------+
```

## How does it work?
* Humans (or workers) submit task specifications to a Colony via a Colonies server, which manage one or several Colonies. 
* Colonies workers connect to the Colonies server and search for suitable tasks to execute. Each worker must have a valid identity (like a passport) to prove its Colony membership and the Colonies server makes sure only authorized and qualified workers can connect and be assigned relevant tasks. 
* Colonies workers can **reside anywhere on the Internet**, e.g. a server, inside a Kubernetes Pod, a smart phone app, or embedded in a web page, thus enabling a compute continuum spanning devices, edge and cloud.
* If a worker fails to complete a task in time, the task will be re-assigned to another worker. This fail-safe mechanism ensures that all tasks are eventually completed. This also makes it possible to apply **Chaos Engineering**, e.g. randomly kill workers to test the overall stability of the system.  

## What is it good at?
* **Distributed computing**. Manage ML/AI workloads on Kubernetes. Form a Colony by deploying one or several Colonies workers in Kubernetes Pods. Then use Colonies to enable batch processing and launch processes inside worker containers.
* **Distributed RPC**. Use Colonies to build overlay networks to manage workflows spanning multiple cloud/edge servers and devices.
* **Grid computing**. Use Colonies as a control server where geographically dispersed workers perform computations.
* **Meta operating systems** for the next-generation IoT, Edge computing and High-Performance Computing (HPC). Use Colonies to integrate various systems together, e.g. a Slurm worker could train a neural network at a super-computer, which are then automatically deployed by another worker to an Edge server or IoT device. Colonies makes it possible to handle these kinds of heterogeneous systems as a single unit to establish a compute continuum across many different systems and platforms.     

## What about Kubernetes and container-native workflow engines?
* Colonies makes it possible to **orchestrate processes inside containers**. This is far more efficient than launching a new container for each new job like [Argo Workflows](https://argoproj.github.io/argo-workflows). This is especially important when dealing with AI workflows consisting of huge containers (tens of gigabytes) or when a large amount of data needs to be shuffled into memory.
* Colonies **complements Kubernetes** and brings robust and fault tolerant **batch processing** to Kubernetes, typically needed by many AI workloads.
* At the same time, Colonies is **lightweight and does not require Kubernetes**. It runs in browsers, smart phones or IoT devices. This also makes it much easier to develop and test complex workflows before they are deployed on Kubernetes.
* Most existing frameworks are not built on top of a crypto-protocol, which makes them hard to use in an overlay across platforms and untrusted networks. 

## Key features
* Colonies is based on [Etcd](https://etcd.io/) and is **scalable** and **robust**. 
* A built-in crypto-protocol ECDSA (Elliptic Curve Digital Signature Algorithm) provides identity management and **secure** and **zero-trust process execution**.
* **Robust batch processing and distributed RPC.** Processes not finishing in time are automatically moved back to the job queue to be executed by another worker.  
* **Pull-based orchestration.** Users (or workers) submit process specifications the Colonies server. Colonies workers connect to the Colonies server and request processes to execute. A HTTP Long Polling/WebSocket protocol ensure that workers can reside anywhere on the Internet, even behind firewalls. The Colonies server never establish connections directly to workers. 
* **Multi-step workflows** or **Directed Acyclic Graph (DAG)** to capture dependencies between jobs.
* **Generators** to automatically spawn new workflows based on external events or timeouts.
* **Traceability**, full process execution history can be stored and used for auditing.

# More information
## Installation
* [Installation](docs/Installation.md)
## Guides
* [Getting started](docs/GettingStarted.md)
* [How to implement a Colonies Worker](docs/Worker.md)
* [How to implement a Fibonacci Worker in Go)](docs/GoTutorial.md)
* [How to create workflows DAGs)](docs/Workflows.md)
* [How to use generators](docs/Generators.md)
* [How to use the Colonies CLI tool](docs/CLI.md)
## Design
* [Overall design](docs/Design.md)
* [HTTP RPC protocol](docs/RPC.md)
* [Security design](docs/Security.md)
## SDKs
* [Golang Colonies SDK](https://github.com/colonyos/colonies/tree/main/pkg/client)
* [Julia Colonies SDK](https://github.com/colonyos/ColonyRuntime.jl)
* [JavaScript Colonies SDK](https://github.com/colonyos/colonyruntime.js)
* [Python Colonies SDK](https://github.com/colonyos/pyruntime)
* [Haskell Colonies SDK](https://github.com/colonyos/haskellrt)
## Deployment
* [High-availability deployment](docs/HADeployment.md)
* [Grafana/Prometheus Monitoring](docs/Monitoring.md)
* [Kubernetes Helm charts](https://github.com/colonyos/helm)

More information can also be found [here](https://colonyos.io).

# Current users
* Colonies is currently being used by **[RockSigma AB](https://www.rocksigma.com)** to build a compute engine for automatic seismic processing in underground mines. 

# Running the tests
Follow the instructions at [Installation Guide](./docs/Installation.md) then type:
```console
make test
```
