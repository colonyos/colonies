## What is Colonies? 
Colonies is a generic framework for implementing next-generation distributed applications and systems. It can for example be used as a building block for implementing an *Edge Computing Operating System* or a "Grid Computing Engine". 

A **Colony** is a collection of (geographically) distributed computers that can be controlled using a single API. A **Colony Runtime** receives intructions from the **Colonies Server** and is responsible for executing processes. The Colonies server works as a mediator, trying to match submitted processes specification to suitable runtimes. It also keep tracks of the history of all process execution and can also re-assign a process to another runtime if it is not completed in time. 

A Colony may consists of many different kinds of Colony Runtimes, e.g. a **Kubernetes Colony Runtime**, **Docker Colony Runtime**, or a **Slurm Colony Runtime**. A Colony Runtime can also reside in IoT devices or smart phones, thus making it possible to deploy and manage applications that run across devices and servers. In this way, Colonies can be used to implement a *Cloud-of-Cloud* platform that combines many execution environments into a new virtual computing environment that can be controlled using an single unified API. 

![Colonies Architecture](docs/images/ColoniesArch.png?raw=true "Colonies Architecture")

### Security principles
A core component of Colonies is a crypto identity protocol inspired by Bitcoin and Ethereum. Each Colony and Colony Runtime is assigned a *Digital Identity* that is verified by the Colonies server using a so-called [Implicit certificates](https://en.wikipedia.org/wiki/Implicit_certificate), which is implemented using [Elliptic-curve cryptography](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography). This protocol makes it possible to reconstruct public-keys from signatures. Identities can then simply be calculated as cryptographic hashes (SHA3-256) of the reconstructed public-keys.

The Colonies Server keeps track of these identities and applies several rules how runtimes are allowed to interact with each other. 

1. Only the Colonies Server Owner may register a new Colony. **Requires rootpassword** specified when starting the Colonies Server. See example below.
2. Only a Colony Owner may register/approve/disapprove/list/get info about Colony Runtimes in a Colony. **Requires a Colony Private key.**
3. Only a Colony Runtime may submit/list/get info about a process. **Requires a Runtime Private Key.**
4. Only a Colony Runtime may set/get/list attributes on a process. **Requires a Runtime Private Key.**

Note that the Colonies server does not store any crypto keys, but rather stores identites in a database and verifies that reconstructed identities obtained from RPC calls matches the identities stored in the database.

## Links
* [Installation](docs/Installation.md)
* [Using the Colonies CLI tool](docs/Cli.md)
