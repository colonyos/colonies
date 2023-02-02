# Security principles
A core component of Colonies is a crypto identity protocol inspired by Bitcoin and Ethereum. Each colony and colony executor is assigned a *Digital Identity* that is verified by the Colonies server using a so-called [Implicit certificates](https://en.wikipedia.org/wiki/Implicit_certificate), which is implemented using [Elliptic-curve cryptography](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography). This protocol makes it possible to reconstruct public-keys from signatures. Identities can then simply be calculated as cryptographic hashes (SHA3-256) of the reconstructed public-keys.

The Colonies server keeps track of these identities and applies several rules how executors are allowed to interact with each other. 

1. Only the server owner can register a new colony. 
2. Only the server owner can list registered colonies. 
3. Only a colony owner can register a colony executor to a colony. 
4. Only a colony owner can list/get info about colony.
5. Only a colony owner can approve/disapprove a executor.
6. Any executor member of a colony can submit/get/assign/list processes or workflows. 
7. Only the executor that was assigned a process can set attributes on that process and close it. 
8. Any executor can get/list attributes on processes. 

Note that the Colonies server does not store any crypto keys, but rather stores identites in a database and verifies that reconstructed identities obtained from RPC calls match the identities stored in the database. This protocol works as follows. Let's assume a executor client has the following Id: 

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
