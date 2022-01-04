[![codecov](https://codecov.io/gh/johankristianss/colonies/branch/main/graph/badge.svg?token=G32O1AO1YB)](https://codecov.io/gh/johankristianss/colonies)

[![Go](https://github.com/johankristianss/colonies/actions/workflows/go.yml/badge.svg)](https://github.com/johankristianss/colonies/actions/workflows/go.yml)

# What is Colonies? 
**Colonies** is a generic framework for implementing next-generation distributed applications and systems. It can be used as a building block for grid computing or edge computing, e.g. implement a *meta operating system* or *cloud-of-cloud* platform that combines many execution environments into a new virtual computing environment that can be controlled using an single unified API.

![Colonies Architecture](docs/images/ColoniesArch.png?raw=true "Colonies Architecture")

* A **Colony** is a trusted collection of (geographically) distributed computers.   
* A **Colony Process** is a virtual software process. It contains meta information how to actually execute a process on a real operating system.
* A **Colony Runtime** communicates with the Colonies **Colonies Server** and provides an virtual runtime environment for executing Colony Processes. 
* The **Colonies Server** works as a mediator, trying to match submitted processes specification to Colony Runtimes. It keep tracks of execution history and can also re-assign a Colony Process if is not completed in time. All Remote Procedure Calls (RPC) to the Colonies server are done atomically. This means that two Runtimes cannot be assigned the same process. If a Runtime crashes during a RPC call, that call is consider invalid, thus keeping the Colonies Server consistent.   
* A **Colony App** is an software that runs somewhere on the Internet. It is a running instance of a Colony Process. A Colony App can for example be an Android app running on a smart phone or an IoT application running a contraint device. A Colony App consists a of Colony Runtime for interacting with other Colony Apps. 
* A **Colony Service** is Colony App that provide services to other Colony Apps or Colony Sevices. For example, a **Colony Kubernetes Service** makes it possible to execute Colony Processes on top of Kubernetes. A **Colony Slurm Services** provides a service to run Colony Processes on HPC supercomputers. Similar to a real-operating system a **Colony App** may spawn new Colony Processes.   

# Links
* [Installation](docs/Installation.md)
* [Using the Colonies CLI tool](docs/CLI.md)
* [Tutorial](docs/Tutorial.md)
* [HTTP RPC Protocol](docs/RPC.md)

# Security principles
A core component of Colonies is a crypto identity protocol inspired by Bitcoin and Ethereum. Each Colony and Colony Runtime is assigned a *Digital Identity* that is verified by the Colonies server using a so-called [Implicit certificates](https://en.wikipedia.org/wiki/Implicit_certificate), which is implemented using [Elliptic-curve cryptography](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography). This protocol makes it possible to reconstruct public-keys from signatures. Identities can then simply be calculated as cryptographic hashes (SHA3-256) of the reconstructed public-keys.

The Colonies Server keeps track of these identities and applies several rules how runtimes are allowed to interact with each other. 

1. Only the Colonies Server Owner can register a new Colony. 
2. Only the Colonies Server Owner can list registered Colonies. 
3. Only a Colony Owner can register a Colony Runtimes to a Colony. 
4. Only a Colony Owner can list/get info about Colony.
5. Only a Colony Owner can approve/disapprove a Runtime.
6. Any Colony Runtime of a Colony can submit/get/list processes. 
7. Only the Colony Runtime that was assigned a process can set attributes on that process. 
8. Any Colony Runtime can get/list attributes on processes. 

Note that the Colonies server does not store any crypto keys, but rather stores identites in a database and verifies that reconstructed identities obtained from RPC calls match the identities stored in the database. This protocol works as follows. Let's assume that a Runtime client has the following Id: 

```
69383f17554afbf81594999eec96adbaa0fc6caace5f07990248b14167c41e8f
```

It then sends the following message to the Colonies Server:

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
 
