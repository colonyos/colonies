# HTTP RPC protocol
The Colonies RPC messages has the following format:

```json
{
    "payloadtype": "addcolonymsg",
    "payload": "ewogICAgICBjb2xvbnlpZDogYWM4ZGM4OTQ5YWYzOTVmZDUxZWFkMzFkNTk4YjI1MmJkYTAyZjFmNmVlZDExYWNlN2ZjN2RjOGRkODVhYzMyZSwKICAgICAgbmFtZTogdGVzdF9jb2xvbnlfbmFtZQogIH0=",
    "signature": "82f2ba6368d5c7d0e9bfa6a01a8fa4d4263113f9eedf235e3a4c7b1febcdc2914fe1f8727746b2f501ceec5736457f218fe3b1a469dd6071775c472a802aa81501",
}
```

* Messages are POSTed to http://host:port/api.
* The *payload* attribute is an Base64 string containing JSON data as specified in the API description below.
* The *signature* is calculated based on the Base64 payload data using a private key.
* It is assumed that SSL/TLS are used to prevent replay attacks.
* Note that **payloadtype** and **msgtype** must match. The reason to duplicate this information is allow for introspection using structured parsning but at the same time sign the message so that the semantic of the RPC operation is kept in one message. Otherwise, an attacker would be able to change the payloadtype and keep the payload to trick the Colonies Server. 

The Colonies Server will reply with a RPC reply message according to the following format:

```json
{
    "payloadtype": "addcolonymsg",
    "payload": "ewogICAgICBjb2xvbnlpZDogNmQ2MWFmZTc5MTRjNjNmMjhhNGM5NzY0NWNlNmFiMjY0YzNhZDNhMGU0NmViZDFmMzc4OGU4MzA1MzkzNGUxOCwKICAgICAgbmFtZTogdGVzdF9jb2xvbnlfbmFtZQogIH0=",
}
```

If the **payloadtype** is set to **error**, then the payload will contain the following JSON data:
```json
{
    "errorcode": "500",
    "msg": "something when wrong here"
}
```

Else it will contain the reply JSON data, e.g:
```json
{
    "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
    "name": "test_colony_name"
}
```

## Colony API

### Add Colony
* PayloadType: **addcolonymsg**
* Credentials: A valid Server Owner Private Key

#### Payload 
```json
{
    "msgtype": "addcolonymsg",
    "colony": {
        "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
        "name": "test_colony_name"
    }
}
```

#### Reply 
```json
{
    "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
    "name": "test_colony_name"
}
```

### List Colonies
* PayloadType: **getcoloniesmsg**
* Credentials: A valid Server Owner Private Key

#### Payload 
```json
{
    "payloadtype": "getcoloniesmsg",
    "payload": "...",
    "signature": "...",
}
```

#### Reply 
```json
[
    {
        "colonyid": "aaae394349008b01c4e56c57a5069aa2e2e8c7e41d9118e04a9039b90b41e93c",
        "name": "test_colony_name"
    },
    {
        "colonyid": "f3127b8c82942e023a8d0b9964203fa00dc22bf7b120e26059d640edeabeb11d",
        "name": "test_colony_name"
    }
]
```

### Get Colony info
* PayloadType: **getcolonymsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "getcolonymsg",
    "colonyid": "42beaae68830094a4b367b06ef293aca0473ae8cd893da43a50000c98c85c5d8"
}
```

#### Reply 
```json
{
    "colonyid": "ac8dc8949af395fd51ead31d598b252bda02f1f6eed11ace7fc7dc8dd85ac32e",
    "name": "test_colony_name"
}
```

## Runtime API
* PayloadType: **addruntimemsg**
* Credentials: A valid Colony Private Key

#### Payload 
```json
{
    "msgtype": "addruntimemsg",
    "runtime": {
        "runtimeid": "38df5bbbcf0ccb438d2e4151638e3967bf28a5654af6a7e5acc590c0e49fae06",
        "runtimetype": "test_runtime_type",
        "name": "test_runtime_name",
        "colonyid": "405acc69052cf19ce23ddd238b73c74bfd78c65cf6ef57613b870470a26d6f95",
        "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
        "cores": 32,
        "mem": 80326,
        "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
        "gpus": 1,
        "state": 0
    }
}
```

#### Reply 
```json
{
    "runtimeid": "38df5bbbcf0ccb438d2e4151638e3967bf28a5654af6a7e5acc590c0e49fae06",
    "runtimetype": "test_runtime_type",
    "name": "test_runtime_name",
    "colonyid": "405acc69052cf19ce23ddd238b73c74bfd78c65cf6ef57613b870470a26d6f95",
    "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
    "cores": 32,
    "mem": 80326,
    "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
    "gpus": 1,
    "state": 0
}
```

### List Runtimes
* PayloadType: **getruntimesmsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "getruntimesmsg",
    "colonyid": "863e313bfd882fe7c0f13c14aff1f3f02ba763bcb48377e50d505289c81e47b6"
}
```

#### Reply 
```json
[
    {
        "runtimeid": "9525365b67efdbbf37bc1fa7628c7e75bafd2f298cd26f75500bc1364b2c4c1c",
        "runtimetype": "test_runtime_type",
        "name": "test_runtime_name",
        "colonyid": "863e313bfd882fe7c0f13c14aff1f3f02ba763bcb48377e50d505289c81e47b6",
        "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
        "cores": 32,
        "mem": 80326,
        "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
        "gpus": 1,
        "state": 1
    }
]
```

### Get Runtime info
* PayloadType: **getruntimemsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "getruntimemsg",
    "runtimeid": "ed2aa78eabe3d1f6fd46ef1247199e9a12faf1a8f1bcba0db51265515c3f08e0"
}
```

#### Reply 
```json
{
    "runtimeid": "ed2aa78eabe3d1f6fd46ef1247199e9a12faf1a8f1bcba0db51265515c3f08e0",
    "runtimetype": "test_runtime_type",
    "name": "test_runtime_name",
    "colonyid": "85ae85e8b6fafddfab1a381ea86a5d7f55e818df6cad8a10e5986d87c57b0683",
    "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
    "cores": 32,
    "mem": 80326,
    "gpu": "NVIDIA GeForce RTX 2080 Ti Rev. A",
    "gpus": 1,
    "state": 2
}
```

### Approve Runtime 
* PayloadType: **approveruntimemsg**
* Credentials: A valid Colony Private Key

#### Payload
```json
{
    "msgtype": "approveruntimemsg",
    "runtimeid": "e40e2862e3a68e1c79af4e9475ef64fbf588e13619f4daa7183673b34e189c87"
}
```

#### Reply
```json
{}
```

### Reject Runtime 
* PayloadType: **rejectruntimemsg**
* Credentials: A valid Colony Private Key

#### Payload 
```json
{
    "msgtype": "rejectruntimemsg",
    "runtimeid": "7804cea6a50f2a258ad815b0ed37b6b312c813bf7387cef04958971335faae21"
}
```

#### Reply 
```json
{}
```

## Process API

### Submit Process Specification 
* PayloadType: **submitprocessespecmsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "submitprocessesspecmsg",
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "2de470e10b87dc261c05f6b2da45d0802044208d6c617a056f4824d958710827",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {
            "test_key_1": "test_value_1"
        }
    }
}
```

#### Reply
```json
{
    "processid": "2c0fd0407292538cb8dce3cb306f88b2ab7f3726d649e07502eb04344d9f7164",
    "assignedruntimeid": "",
    "isassigned": false,
    "state": 0,
    "submissiontime": "2022-01-02T11:58:30.017857Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": [
        {
            "attributeid": "ac17247ca031ea6581617de1083f5f4109756ca2f06a65beecf8fb188e870034",
            "targetid": "2c0fd0407292538cb8dce3cb306f88b2ab7f3726d649e07502eb04344d9f7164",
            "attributetype": 4,
            "key": "test_key_1",
            "value": "test_value_1"
        }
    ],
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "2de470e10b87dc261c05f6b2da45d0802044208d6c617a056f4824d958710827",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {
            "test_key_1": "test_value_1"
        }
    }
}
```

### Assign Process to a Runtime 
* PayloadType: **assignprocessmsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "assignprocessmsg",
    "colonyid": "326691e2b5fc0651b5d781393c7279ab3dc58c6627d0a7b2a09e9aa0e4a60950"
}
```

#### Reply 
```json
{
    "processid": "68db01b27271168cb1011c1c54cc31a54f23eb7e5767e49bb34fb206591d2a65",
    "assignedruntimeid": "d02274979e69d534202ca4cdcb3847c56e860d09039399feee6358b8c285d502",
    "isassigned": true,
    "state": 1,
    "submissiontime": "2022-01-02T12:01:41.751942Z",
    "starttime": "2022-01-02T12:01:41.756226473+01:00",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": null,
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "326691e2b5fc0651b5d781393c7279ab3dc58c6627d0a7b2a09e9aa0e4a60950",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### List processes
* PayloadType: **getprocessesmsg**
* Credentials: A valid Runtime Private Key

#### Payload 
The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

```json
{
    "msgtype": "getprocessesmsg",
    "coloyid": "891f0c88e8a00cb103df472e4ece347a41eb0115e5c40f12d565bb24eb3fc71d",
    "count": 2,
    "state": 3 
}
```

#### Reply 
```json
[
    {
        "processid": "88169d23b0828ed65f0a007e4be6bf9734358b9a64379d0c6e53a0496216db4c",
        "assignedruntimeid": "653c818113e878d704935e639371f72a3167d510008607c70176e8147adf7865",
        "isassigned": true,
        "state": 3,
        "submissiontime": "2022-01-02T12:04:21.647969Z",
        "starttime": "2022-01-02T12:04:21.657305Z",
        "endtime": "2022-01-02T12:04:21.661402Z",
        "deadline": "0001-01-01T00:00:00Z",
        "retries": 0,
        "attributes": null,
        "spec": {
            "timeout": -1,
            "maxretries": 3,
            "conditions": {
                "colonyid": "891f0c88e8a00cb103df472e4ece347a41eb0115e5c40f12d565bb24eb3fc71d",
                "runtimeids": [],
                "runtimetype": "test_runtime_type",
                "mem": 1000,
                "cores": 10,
                "gpus": 1
            },
            "env": {}
        }
    }
]
```

### Get Process info
* PayloadType: **getprocessmsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "getprocessmsg",
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce"
}
```

#### Reply 
```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedruntimeid": "",
    "isassigned": false,
    "state": 0,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": null,
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "ee193a3f4f3f93bfc87801cf1d01511c12c199cb80bfbf4955bb3d9d4638720d",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### Close Process as Successful 
* PayloadType: **closesuccessfulmsg**
* Credentials: A valid Runtime Private Key and the Runtime ID needs to match the RuntimeID assigned to the process

#### Payload 
```json
{
    "msgtype": "closesuccessfulmsg",
    "processid": "ed041355071d2ee6d0ec27b480e2e4c8006cf465ec408b57fcdaa5dac76af8e2"
}
```

#### Reply
```json
{}
```

### Close a Proceess as Failed 
* PayloadType: **closefailedmsg**
* Credentials: A valid Runtime Private Key and the Runtime ID needs to match the RuntimeID assigned to the process

#### Payload 
```json
{
    "msgtype": "closefailedfulmsg",
    "processid": "24f6d85804e2abde0c85a9e8aef8b308c44a72323565b14f11756d4997acf200"
}
```

#### Reply 
```json
{}
```

### Add Attribute to a Process 
* PayloadType: **addattributemsg**
* Credentials: A valid Runtime Private Key and the Runtime ID needs to match the RuntimeID assigned to the process

#### Payload 
```json
{
    "msgtype": "addattributemsg",
    "attribute": {
        "attributeid": "216e26cb089032d2f941454e7db5f3ae1591eeb43eb477c3f8ed545b96d4f690",
        "targetid": "c4775cab695da8a77b503bbe29df8ae39dafd1c7fed3275dac11b436c1724dbf",
        "attributetype": 1,
        "key": "result",
        "value": "helloworld"
    }
}
```

#### Reply 
```json
{
    "attributeid": "216e26cb089032d2f941454e7db5f3ae1591eeb43eb477c3f8ed545b96d4f690",
    "targetid": "c4775cab695da8a77b503bbe29df8ae39dafd1c7fed3275dac11b436c1724dbf",
    "attributetype": 1,
    "key": "result",
    "value": "helloworld"
}
```

### Get Attribute assigned to a Process 
* PayloadType: **getattributemsg**
* Credentials: A valid Runtime Private Key

#### Payload 
```json
{
    "msgtype": "getattributemsg",
    "attributeid": "a1d8f3613e074a250c2fbab478a0e11eb40defee66bd9b6a6ceb96990f1486eb"
}
```

#### Reply 
```json
{
    "attributeid": "a1d8f3613e074a250c2fbab478a0e11eb40defee66bd9b6a6ceb96990f1486eb",
    "targetid": "3d893a44a30c7e5c5c595413a9de1545a9d43a844528831c4e205b280c074e56",
    "attributetype": 1,
    "key": "result",
    "value": "helloworld"
}
```

### Subscribe Process Events
* PayloadType: **subscribeprocessmsg**
* Credentials: A valid Runtime Private Key
* Comments: Receives an event when a process changes state. The payload needs to be sent over a websocket to: wss://host:port/pubsub

#### Payload 
The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

```json
{
    "msgtype": "subscribeprocessmsg",
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "state": 1,
    "timeout": -1
}
```

#### Reply 
```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedruntimeid": "",
    "isassigned": false,
    "state": 0,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": null,
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "ee193a3f4f3f93bfc87801cf1d01511c12c199cb80bfbf4955bb3d9d4638720d",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### Subscribe Processes Events
* PayloadType: **subscribeprocessesmsg**
* Credentials: A valid Runtime Private Key
* Comments: Receives an event when processes are added or change state. The payload needs to be sent over a websocket to: wss://host:port/pubsub

#### Payload 
The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

```json
{
    "msgtype": "subscribeprocessesmsg",
    "runtimetype": "test_runtime_type",
    "state": 1,
    "timeout": -1
}
```

#### Reply 
```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedruntimeid": "",
    "isassigned": false,
    "state": 0,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": null,
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "ee193a3f4f3f93bfc87801cf1d01511c12c199cb80bfbf4955bb3d9d4638720d",
            "runtimeids": [],
            "runtimetype": "test_runtime_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```
