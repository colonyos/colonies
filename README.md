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

## Guide
### Register a new Colony
First, create a file named colony.json, and put the following content into it.
```
{
    "name": "mycolony"
}
```

Then use the colonies tool to register the colony. The id of the colony will be returned if the command is successful. Note that the root password is required for this operation.
```
$ colonies colony register --rootpassword=secret --json ./examples/colony.json 
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
```
$ colonies computer register --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --json computer.json
```
