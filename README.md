## What is Colonies? 
Colonies is a generic framework for implementing next-generation distributed applications and systems. It can for example be used as a building block for implementing an *Edge Computing Operating System* or a "Grid Computing Engine". 

A **Colony** is a collection of (geographically) distributed computers that can be controlled using a single API. A **Colony Runtime** receives intructions from the **Colonies Server** and is responsible for executing processes. The Colonies server works as a mediator, trying to match submitted processes specification to suitable runtimes. It also keep tracks of the history of all process execution and can also re-assign a process to another runtime if it is not completed in time. 

A Colony may consists of many different kinds of Colony Runtimes, e.g. a **Kubernetes Colony Runtime**, **Docker Colony Runtime**, or a **Slurm Colony Runtime**. A Colony Runtime can also reside in IoT devices or smart phones, thus making it possible to deploy and manage applications that run across devices and servers. In this way, Colonies can be used to implement a *Cloud-of-Cloud* platform that combines many execution environments into a new virtual computing environment that can be controlled using an single unified API. 

![Colonies Architecture](docs/images/ColoniesArch.png?raw=true "Colonies Architecture")

### Security principles
A core component of Colonies is a crypto identity protocol inspired by Bitcoin and Ethereum. Each Colony and Colony Runtime is assigned a *Digital Identity* that is verified by the Colonies server using a so-called [Implicit certificates](https://en.wikipedia.org/wiki/Implicit_certificate), which is implemented using [Elliptic-curve cryptography](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography). This protocol makes it possible to reconstruct public-keys from signatures. Identities can then simply be calculated as cryptographic hashes (SHA3-256) of the reconstructed public-keys.

The Colonies Server keeps track of these identities and applies several rules how runtimes are allowed to interact with each other. 

1. Only the Colonies Server Owner can register a new Colony. 
   - Credentials: Requires a valid Server Private Key.
3. Only the Colonies Server Owner can list registered Colonies. 
   - Credentials: Requires a valid Server Private Key.
5. Only a Colony Owner can register a Colony Runtimes to a Colony. 
   - Credentials: Requires a valid Colony Private key.
7. Only a Colony Owner can approve/disapprove a Colony Runtimes member of a Colony. 
   - Credentials: Requires a valid Colony Private key.
9. Only a Colony Owner can list/get info about Colony Runtimes member of a Colony. 
   - Credentials Requires a valid Colony Private key.
11. Any Colony Runtime member of a Colony can submit/get/list processes. 
   - Credentials: Requires a valid Runtime Private Key.
13. Only the Colony Runtime that was assigned a process can set attributes on that process. 
   - Credentials: Requires a valid Runtime Private Key.
15. Any Colony Runtime can get/list attributes on processes. 
   - Credentials: Requires a valid Runtime Private Key.

Note that the Colonies server does not store any crypto keys, but rather stores identites in a database and verifies that reconstructed identities obtained from RPC calls match the identities stored in the database. This protocol works as follows. Let's assume that a Runtime client has the following Id: 

```
69383f17554afbf81594999eec96adbaa0fc6caace5f07990248b14167c41e8f
```

It then sends the following message to the Colonies Server:

```json
{
    "rpc": {
        "method": "GetProcess",
        "nonce": "3473e116d839228cf38f964392520e97af426ebc8cc6a0d1b708be05ded5eef9"
    },
    "processid": "bba9fc3a63cde17c1fb96f246873c34048fb6d47f2d0aab351d487dc2b30e0e7"
}
```

Additionally, the client also generate a signature using the client's private key and sends the signature togeather with the message to the server.
```
fddcb99aa2ced69771a4d177db0bf9449add1b82d4d41da3c6ef50f56cb487de17f9ab10835e4b23c3981c67382852e7eea2f28708105e06b7c19ec54032ad0001
```

When receiving the message, the server reconstructs the Id of the calling client using the received signature and message. This means that client Id (e.g. 69383f17554afbf815...) is never sent to the server but rather derived by the server from the message it receives. The server will now check in the database if the reconstructed Runtime Id is a member of the colony where the requested process is running.

## Links
* [Installation](docs/Installation.md)
* [Using the Colonies CLI tool](docs/Cli.md)
