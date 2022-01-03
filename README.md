[![codecov](https://codecov.io/gh/johankristianss/colonies/branch/main/graph/badge.svg?token=G32O1AO1YB)](https://codecov.io/gh/johankristianss/colonies)

# What is Colonies? 
**Colonies** is a generic framework for implementing next-generation distributed applications and systems. It can be used as a building block for grid computing or edge computing (e.g. implement a meta operating system).

A **Colony** is a collection of (geographically) distributed computers that can be controlled using a single API. A **Colony Runtime** receives intructions from the **Colonies Server** and is responsible for executing processes. The Colonies server works as a mediator, trying to match submitted processes specification to runtimes. It keep tracks of process execution history and can re-assign processes if a runtime does not complete a process in time. All Remote Procedure Calls (RPC) to the Colonies server are done atomically. This means that two Runtimes cannot be assigned the same process. If a Runtime crashes during a RPC call, that call is consider invalid, thus keeping the database consistent.   

A Colony may consists of many different kinds of Colony Runtimes, e.g. a **Kubernetes Colony Runtime**, **Docker Colony Runtime**, or a **Slurm Colony Runtime**. A Colony Runtime can also reside in IoT devices or smart phones, thus making it possible to deploy and manage applications that run across devices and servers. In this way, Colonies can be used to implement a *Cloud-of-Cloud* platform that combines many execution environments into a new virtual computing environment that can be controlled using an single unified API.

![Colonies Architecture](docs/images/ColoniesArch.png?raw=true "Colonies Architecture")

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
    "method": "addcolony",
    "signature": "82f2ba6368d5c7d0e9bfa6a01a8fa4d4263113f9eedf235e3a4c7b1febcdc2914fe1f8727746b2f501ceec5736457f218fe3b1a469dd6071775c472a802aa81501",
    "payload": "ewogICAgICBjb2xvbnlpZDogYWM4ZGM4OTQ5YWYzOTVmZDUxZWFkMzFkNTk4YjI1MmJkYTAyZjFmNmVlZDExYWNlN2ZjN2RjOGRkODVhYzMyZSwKICAgICAgbmFtZTogdGVzdF9jb2xvbnlfbmFtZQogIH0=",
    "error": false
}
```

When the server receives the message, it reconstructs the Id of the calling client using the enclosed signature and payload. This means that client Id (e.g. 82f2ba6368d5c7d0e9bfa6...) is never sent to the server but rather derived by the server from messages it receives. In the example above, the server checks in the database if the reconstructed Id is a server owner.

# Running the Tests
Follow the instructions at [Installation Guide](./docs/Installation.md) and setup a Postgresql server, then type:
```console
make test
```json
 
