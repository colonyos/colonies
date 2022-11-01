[![codecov](https://codecov.io/gh/colonyos/colonies/branch/main/graph/badge.svg?token=1D4O2JVSJL)](https://codecov.io/gh/colonyos/colonies)
[![Go](https://github.com/colonyos/colonies/actions/workflows/go.yml/badge.svg)](https://github.com/colonyos/colonies/actions/workflows/go.yml)

![ColonyOSLogo](docs/images/ColonyOsLogoNoShaddow2.png)

# What is Colonies?
Colonies is a **Process Orchestration** framework for managing AI/ML workloads across heterogeneous computing platforms. 
    
Colonies provides functionality to establish trusted distributed computing environments, process management, and functionality to orchestrate workloads. It is a building block to implement a [Meta-Operating System](https://en.wikipedia.org/wiki/Meta-system"), an overlay built on top of existing operating systems and platforms to create *compute continuums spanning devices, webapps, clouds, and edge and HPC platforms*.

The core idea of Colonies is to split up complex workloads in two layers, a **Meta-layer** and an **Execution-layer**.

![MetaOS](docs/images/meta-os.png)

* The **Meta-layer** makes it possible to describe and manage complex workflows as **meta-processes** independently of implementation and execution environment.
* The **Execution-layer** provides a serverless computing environment where developers can implement workers capable of executing certain types of meta-processes. AI applications can then be broken down into composable functions executed by remote workers anywhere on the Internet.
* A build-in [zero-trust](https://en.wikipedia.org/wiki/Zero_trust_security_model) protocol makes it possible to organize remote workers as a single unit called a **Colony**, thus making it possible for users to keep control even if workloads are spread out and executed on many different platforms at the same time. 

Below are some screenshots from the [Colonies Dashboard](https://github.com/colonyos/dashboard)
![Dashboard](docs/images/dashboard1.png)
![Dashboard](docs/images/dashboard2.png)
![Dashboard](docs/images/dashboard3.png)

# More information
## Installation
* [Installation](docs/Installation.md)
## Presentations
* [Process Orchestration with ColonyOS](docs/Colonies.pptx)
## Guides
* [Introduction](docs/Introduction.md)
* [Getting started](docs/GettingStarted.md)
* [How to implement a Colonies worker](docs/Worker.md)
* [How to implement a FibonacciwWorker in Go](docs/GoTutorial.md)
* [How to create workflows DAGs](docs/Workflows.md)
* [How to use generators](docs/Generators.md)
* [How to use crons](docs/Crons.md)
* [How to use the Colonies CLI](docs/CLI.md)
## Design
* [Overall design](docs/Design.md)
* [HTTP RPC protocol](docs/RPC.md)
* [Security design](docs/Security.md)
## SDKs
* [Golang Colonies SDK](https://github.com/colonyos/colonies/tree/main/pkg/client)
* [Rust Colonies SDK](https://github.com/colonyos/rustrt)
* [Julia Colonies SDK](https://github.com/colonyos/ColonyRuntime.jl)
* [JavaScript Colonies SDK](https://github.com/colonyos/colonyruntime.js)
* [Python Colonies SDK](https://github.com/colonyos/pyruntime)
* [Haskell Colonies SDK](https://github.com/colonyos/haskellrt)
## Deployment
* [High-availability deployment](docs/HADeployment.md)
* [Grafana/Prometheus monitoring](docs/Monitoring.md)
* [Kubernetes Helm charts](https://github.com/colonyos/helm)

More information can also be found [here](https://colonyos.io).

# Current users
* Colonies is currently being used by **[RockSigma AB](https://www.rocksigma.com)** to build a compute engine for automatic seismic processing in underground mines. 

# Running the tests
Follow the instructions at [Installation Guide](./docs/Installation.md) then type:
```console
make test
```
