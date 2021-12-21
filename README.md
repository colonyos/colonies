## What is Colonies? 
Colonies is a generic framework for implementing distributed applications and systems. It can for example be used for implementing on Edge Computing Operating System.

## Getting started
## Installation
### Start a TimescaleDB server
```
$ docker run -d --name timescaledb -p 5432:5432 -v /storage/fast/lib/timescaledb/data:/var/lib/postgresql/data -e POSTGRES_PASSWORD=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7 --restart unless-stopped timescale/timescaledb:latest-pg12
```

### Setup a database
```
$ colonies database create --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### In case, you would like to clear the database
```
$ colonies database drop --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### Start a Colonies server 
```
$ colonies server start --rootpassword=secret --port=8080 --tlscert=./cert/cert.pem --tlskey=./cert/key.pem --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

## Using the Colonies CLI tool 
### Register a new Colony
First, create a file named colony.json, and put the following content into it.
```
{
    "name": "mycolony"
}
```

Then use the colonies tool to register the colony. The id of the colony will be returned if the command is successful. Note that the root password is required for this operation.
```
$ colonies colony register --rootpassword=secret --spec ./examples/colony.json 
2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773
```

### List all Colonies 
Note that root password of Colonies server is also required to list all colonies.
```
$ colonies colony ls --rootpassword=secret
[
    {
        "colonyid": "2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773",
        "name": "mycolony"
    }
]
```

### Get the private key of a Colony (or a Computer)
```
$ colonies keychain privatekey --id 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 
4b24941ca1d85fb1ff055e81fad7dba97471e756bebc38e03e657c738f0e1224
```

### Register a new Colony Computer
Only the colony owner is allowed to register a new Colony computer. 
```
$ colonies computer register --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --spec ./examples/computer.json
4c60e0e108690dc034a3f3c6e369e63e077aa4c9795cf46c531938efc4e67243
```

The private key for the colony owner is automatically obtained from the keychain. It is also possible to specify the 
private key as an argument. 

```
$ colonies computer register --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --colonyprvkey=4b24941ca1d85fb1ff055e81fad7dba97471e756bebc38e03e657c738f0e1224 --spec computer.json
```

### List registered Colony Computers
```
$ colonies computer ls --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773
[
    {
        "computerid": "2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd389",
        "name": "my_computer",
        "colonyid": "2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773",
        "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
        "cores": 32,
        "mem": 80326,
        "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
        "gpus": 1,
        "status": 0
    }
]
```

The private key for the colony owner is automatically obtained from the keychain. It is also possible to specify the 
private key as an argument, as in the example above. 

### Approve Colony Computers 
A Colony Computer needs to be approved by the Colony owner before it can execute processes. As before, the private key is automatically fetched from the keychain.
```
$ colonies computer approve --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --computerid 2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd389
```

### Disapprove Colony Computers 
Similarly, a Colony Computer can be disapproved with the "disapprove" command.
```
$ colonies computer disapprove --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --computerid 2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd389
```

### Submit a process to a Colony 
First we need to create a process spec file.
```
{
    "computertype": "test_computer_type",
    "timeout": -1,
    "retries": 0,
    "maxretries": 3,
    "mem": 1000,
    "cores": 10,
    "gpus": 1,
}
```
And a input file. 
```
{
    "cmd": "helloworld",
    "args": "hello"
}
```

```
$ colonies process submit --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --spec ./examples/process.json -in ./examples/input.js
```
