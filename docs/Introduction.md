# Introduction

## How does it work?
* Humans (or workers) submit process specs to a Colony via a Colonies server.
* Colonies workers connect to the Colonies server and **search for suitable tasks to execute**. Each worker must have a **valid identity** (like a passport) to prove its Colony membership and the Colonies server makes sure only authorized and qualified workers can connect and be assigned processes. 
* Colonies workers can **reside anywhere on the Internet**, e.g. a server, inside a Kubernetes Pod, a smart phone app, or embedded in a web page, thus enabling a compute continuum spanning devices, edge and cloud.
* If a worker fails to complete a task in time, the task **will be re-assigned to another worker**. This fail-safe mechanism ensures that all tasks are eventually completed. This also makes it possible to apply **Chaos Engineering**, e.g. randomly kill workers to test the overall stability of the system.  

## What is it good at?
* **Distributed computing**. Manage ML/AI workloads on Kubernetes. Form a Colony by deploying one or several Colonies workers in Kubernetes Pods. Then use Colonies to enable batch processing and launch processes inside worker containers.
* **Distributed RPC**. Use Colonies to build overlay networks to manage workflows spanning multiple cloud/edge servers and devices.
* **Grid computing**. Use Colonies as a control server where geographically dispersed workers perform computations.
* **Serverless computing**. Use Colonies as a building block for serverless computing.
* **Meta operating systems**. Use Colonies to integrate various systems together, e.g. a Slurm worker could train a neural network at a super-computer, which are then automatically deployed by another worker to an Edge server or IoT device. Colonies makes it possible to handle these kinds of heterogeneous systems as a single unit to establish a compute continuum across many different systems and platforms.     

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
