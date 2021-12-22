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
$ ./bin/colonies database create --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### In case, you would like to clear the database
```
$ ./bin/colonies database drop --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### Start a Colonies server 
```
$ ./bin/colonies server start --rootpassword=secret --port=8080 --tlscert=./cert/cert.pem --tlskey=./cert/key.pem --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
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
$ ./bin/colonies colony register --rootpassword=secret --spec ./examples/colony.json 
2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773
```

### List all Colonies 
Note that root password of Colonies server is also required to list all colonies.
```
$ ./bin/colonies colony ls --rootpassword=secret
```

```json
[
    {
        "colonyid": "2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773",
        "name": "mycolony"
    }
]
```

### Get the private key of a Colony (or a Computer)
```
$ ./bin/colonies keychain privatekey --id 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 
4b24941ca1d85fb1ff055e81fad7dba97471e756bebc38e03e657c738f0e1224
```

### Register a new Colony Computer
Only the colony owner is allowed to register a new Colony computer. 
```
$ ./bin/colonies computer register --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --spec ./examples/computer.json
4c60e0e108690dc034a3f3c6e369e63e077aa4c9795cf46c531938efc4e67243
```

The private key for the colony owner is automatically obtained from the keychain. It is also possible to specify the 
private key as an argument. 

```
$ ./bin/colonies computer register --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --colonyprvkey=4b24941ca1d85fb1ff055e81fad7dba97471e756bebc38e03e657c738f0e1224 --spec ./examples/computer.json
```

### List registered Colony Computers
```
$ ./bin/colonies computer ls --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773
```

```json
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
$ ./bin/colonies computer approve --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --computerid 2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd389
```

### Disapprove Colony Computers 
Similarly, a Colony Computer can be disapproved with the "disapprove" command.
```
$ ./bin/colonies computer disapprove --colonyid 2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773 --computerid 2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd389
```

### Submit a process to a Colony 
First we need to create a process spec file.

```json
{
    "targetcolonyid": "2770116b0d66a71840a4513bec52707c4a26042462b62e0830497724f7d37773",
    "targetcomputerids": [],
    "computertype": "test_computer_type",
    "timeout": -1,
    "maxretries": 0,
    "mem": 1000,
    "cores": 10,
    "gpus": 1,
    "in": {
        "test_key1": "test_value1",
        "test_key2": "test_value2"
    }
}
```

To submit the process spec to the Colony, type:

```
$ ./bin/colonies process submit --id 2089d3897e512a4e16cfb99d781cb494044323216ec6a1fffecb4da4312fd38 --spec ./examples/process_spec.json
```

```json
{
    "processid": "817f0353e55ea9fa41ebc6d92622bbba49da5ff521751fb0136f99530e2e1d76",
    "targetcolonyid": "5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4",
    "targetcomputerids": [],
    "assignedcomputerid": "",
    "status": 0,
    "isassigned": false,
    "computertype": "test_computer_type",
    "submissiontime": "2021-12-21T21:04:07.807378Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "timeout": -1,
    "retries": 0,
    "maxretries": 0,
    "mem": 1000,
    "cores": 10,
    "gpus": 1,
    "in": [
        {
            "attributeid": "285aef923eaf1830c295ee185a695202f5b2b83746aae3d92ffde7caad4d8253",
            "targetid": "817f0353e55ea9fa41ebc6d92622bbba49da5ff521751fb0136f99530e2e1d76",
            "attributetype": 1,
            "key": "test_key1",
            "value": "test_value1"
        },
        {
            "attributeid": "c51af5ed05a14db70fc64c30f62a71653709817f840270f5b6755894659eb106",
            "targetid": "817f0353e55ea9fa41ebc6d92622bbba49da5ff521751fb0136f99530e2e1d76",
            "attributetype": 1,
            "key": "test_key2",
            "value": "test_value2"
        }
    ],
    "out": null,
    "err": null
} 
```

### List all waiting/submitted processes for a certain computer 
The first waiting processes has highest priority.

```
$ ./bin/colonies process psw --computerid d7f4e767f4efd1b78c7f129c62610b622168b4f69400bcc3bec7b72eeeb4e7bc --colonyid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4
```

```json
[
    {
        "processid": "d9f987677edd9a88ef95c48ceb1ffc76008e4050a8e95cba8212e65599c5b735",
        "targetcolonyid": "5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4",
        "targetcomputerids": [],
        "assignedcomputerid": "",
        "status": 0,
        "isassigned": false,
        "computertype": "test_computer_type",
        "submissiontime": "2021-12-21T20:59:23.563125Z",
        "starttime": "0001-01-01T00:00:00Z",
        "endtime": "0001-01-01T00:00:00Z",
        "deadline": "0001-01-01T00:00:00Z",
        "timeout": -1,
        "retries": 0,
        "maxretries": 0,
        "mem": 1000,
        "cores": 10,
        "gpus": 1,
        "in": null,
        "out": [
            {
                "attributeid": "9b9038a5bb3966daf3684d2f5fa4d886eae66a4222eae7948eb87ba6905e1b69",
                "targetid": "d9f987677edd9a88ef95c48ceb1ffc76008e4050a8e95cba8212e65599c5b735",
                "attributetype": 1,
                "key": "test_key1",
                "value": "test_value1"
            },
            {
                "attributeid": "a3f99acddbd186f9917e00c751189b679ef4d6c92ec5dc0cd72c78ec8645ff17",
                "targetid": "d9f987677edd9a88ef95c48ceb1ffc76008e4050a8e95cba8212e65599c5b735",
                "attributetype": 1,
                "key": "test_key2",
                "value": "test_value2"
            }
        ],
        "err": null
    },
```

### Assign a process to a computer 
```
$ ./bin/colonies process assign --colonyid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4 --computerid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4
```

```json
{
    "processid": "a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177",
    "targetcolonyid": "7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82",
    "targetcomputerids": [],
    "assignedcomputerid": "163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95",
    "status": 1,
    "isassigned": true,
    "computertype": "test_computer_type",
    "submissiontime": "2021-12-22T08:23:44.158115Z",
    "starttime": "2021-12-22T08:23:45.369808389+01:00",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "timeout": -1,
    "retries": 0,
    "maxretries": 0,
    "mem": 1000,
    "cores": 10,
    "gpus": 1,
    "in": [
        {
            "attributeid": "c32a53097bbbfb0245238dfa0c04acc2b66f662bac63647a0ef5b6c065d32bd4",
            "targetid": "a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177",
            "attributetype": 0,
            "key": "test_key1",
            "value": "test_value1"
        },
        {
            "attributeid": "94e02a0c808127f02141acd3838ccdcc12066ae2fd13be55b034a00a339d9d2d",
            "targetid": "a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177",
            "attributetype": 0,
            "key": "test_key2",
            "value": "test_value2"
        }
    ],
    "out": null,
    "err": null
}
```

### List all running processes
The id flag can either be Colony owner or a Computer id.
```
$ ./bin/colonies process ps --colonyid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4 --id 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4
```

```json
[
    {
        "processid": "8c3016c897be1f1c0cf9976b98ffa119076a92c8f8f4a8087d912d354e878a10",
        "targetcolonyid": "7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82",
        "targetcomputerids": [],
        "assignedcomputerid": "163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95",
        "status": 1,
        "isassigned": true,
        "computertype": "test_computer_type",
        "submissiontime": "2021-12-22T08:13:34.135203Z",
        "starttime": "2021-12-22T08:15:15.564647Z",
        "endtime": "0001-01-01T00:00:00Z",
        "deadline": "0001-01-01T00:00:00Z",
        "timeout": -1,
        "retries": 0,
        "maxretries": 0,
        "mem": 1000,
        "cores": 10,
        "gpus": 1,
        "in": [
            {
                "attributeid": "f66390a7bebb0e5333e5f58de51f97c7a20e08a4b8b65f1c6cb93a30dfc50f2b",
                "targetid": "8c3016c897be1f1c0cf9976b98ffa119076a92c8f8f4a8087d912d354e878a10",
                "attributetype": 0,
                "key": "test_key2",
                "value": "test_value2"
            },
            {
                "attributeid": "23478ec3946d33a5955358034182e7d02f008e96fa5b5c2f40a7233f3cf5a434",
                "targetid": "8c3016c897be1f1c0cf9976b98ffa119076a92c8f8f4a8087d912d354e878a10",
                "attributetype": 0,
                "key": "test_key1",
                "value": "test_value1"
            }
        ],
        "out": null,
        "err": null
    },
```

### List all successful processes
The id flag can either be Colony owner or a Computer id.
```
$ ./bin/colonies process pss --colonyid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4 --id 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4
No successful processs found
```

### List all failed processes
The id flag can either be Colony owner or a Computer id.
```
$ ./bin/colonies process psf --colonyid 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4 --id 5dd23ec3bf9d643d47eb8486845071dcf0cfcba4362c3a541ea7cfea5174b7d4
No failed processs found
```

### Get info about a process
```
./bin/colonies process get --processid b5b1b347888414da99c971d9266640429a2ee6d94e89bf83ecac89dd4d3438df  --colonyid 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82 --id 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82
```

```json
{
    "processid": "b5b1b347888414da99c971d9266640429a2ee6d94e89bf83ecac89dd4d3438df",
    "targetcolonyid": "7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82",
    "targetcomputerids": [],
    "assignedcomputerid": "163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95",
    "status": 1,
    "isassigned": true,
    "computertype": "test_computer_type",
    "submissiontime": "2021-12-22T08:22:31.505733Z",
    "starttime": "2021-12-22T08:22:33.18657Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "timeout": -1,
    "retries": 0,
    "maxretries": 0,
    "mem": 1000,
    "cores": 10,
    "gpus": 1,
    "in": [
        {
            "attributeid": "61a914382197d1feca547ec5a73f6e77c25ca3e4cb2a2ee45046e85229e4db6b",
            "targetid": "b5b1b347888414da99c971d9266640429a2ee6d94e89bf83ecac89dd4d3438df",
            "attributetype": 0,
            "key": "test_key1",
            "value": "test_value1"
        },
        {
            "attributeid": "e733b51449a3f85b383827b968b9374450d9ae52ef6a33841507db3a0fdd4d45",
            "targetid": "b5b1b347888414da99c971d9266640429a2ee6d94e89bf83ecac89dd4d3438df",
            "attributetype": 0,
            "key": "test_key2",
            "value": "test_value2"
        }
    ],
    "out": null,
    "err": null
}
```


### Add attribute to a running process 
```
./bin/colonies attribute add --key test_xcdd --value sdsdsdd --colonyid 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82 --computerid 163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95 --processid a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177
```

```json
{
    "attributeid": "f9d8ffd1a156b78c7ebd47fe38e1a466b8ce21630b35401f476941a5e12c5e44",
    "targetid": "a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177",
    "attributetype": 1,
    "key": "test_xssdddddscdd",
    "value": "sdsdsddd"
} 
```

### Get attribute of a running process 
```
./bin/colonies attribute get --attributeid f9d8ffd1a156b78c7ebd47fe38e1a466b8ce21630b35401f476941a5e12c5e44 --colonyid 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82 --processid a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177 --id 163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95
```

```json
{
    "attributeid": "f9d8ffd1a156b78c7ebd47fe38e1a466b8ce21630b35401f476941a5e12c5e44",
    "targetid": "a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177",
    "attributetype": 1,
    "key": "test_xssdddddscdd",
    "value": "sdsdsddd"
} 
```

### Mark a process as successful
```
./bin/colonies process successful --colonyid 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82 --processid a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177 --computerid 163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95
Process marked as successful
```

### Mark a process as failed 
```
./bin/colonies process failed --colonyid 7c7b3582fd05fda2e39ac70c7c6a214be735090eeb2d8db636a4ad4424dcca82 --processid a5620e355153765ed52c4068aff8d17bf617e4d7d3167ded3ccd3fff157a4177 --computerid 163a086a2dcb21de28144379cfa3fb0bd4a64ae06956ee816c2a52b999b00c95
Process marked as successful
```


