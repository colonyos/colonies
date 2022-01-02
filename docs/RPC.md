# HTTP RPC protocol
The Colonies RPC messages has the following format:

```json
{
    "method": "addcolony",
    "signature": "82f2ba6368d5c7d0e9bfa6a01a8fa4d4263113f9eedf235e3a4c7b1febcdc2914fe1f8727746b2f501ceec5736457f218fe3b1a469dd6071775c472a802aa81501",
    "payload": "ewogICAgICBjb2xvbnlpZDogYWM4ZGM4OTQ5YWYzOTVmZDUxZWFkMzFkNTk4YjI1MmJkYTAyZjFmNmVlZDExYWNlN2ZjN2RjOGRkODVhYzMyZSwKICAgICAgbmFtZTogdGVzdF9jb2xvbnlfbmFtZQogIH0=",
    "error": false
}
```

* Messages are POSTed to http://host:port/api.
* The *payload* attribute is an Base64 string containing JSON data as specified in the API description below.
* The *signature* is calculated based on the Base64 payload data using a private key.
* If the *error* attribute is true, then the payload will contain the following JSON data.
* It assumed that SSL/TLS are used to prevent replay attacks.

```json
{
    "errorcode": "500",
    "msg": "something when wrong here"
}
```

## Colony API
### Add Colony

#### RPC Message 
Needs to be signed by a valid Server Owner Private Key.

```json
{
    "method": "addcolony",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload 

```json
{
    "colony": {
        "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
        "name": "test_colony_name"
    }
}
```

#### Decoded reply 

```json
{
    "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
    "name": "test_colony_name"
}
```

### List Colonies

#### RPC Message 
Needs to be signed by a valid Server Owner Private Key. Note that the message is empty except for the timestamp field.

```json
{
    "method": "getcolonies",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
  "timestamp": XXXXX
}
```

#### Decoded reply

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

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "getcolony",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "colonyid": "42beaae68830094a4b367b06ef293aca0473ae8cd893da43a50000c98c85c5d8"
}
```

#### Decoded reply

```json
{
    "colonyid": "ac8dc8949af395fd51ead31d598b252bda02f1f6eed11ace7fc7dc8dd85ac32e",
    "name": "test_colony_name"
}
```

## Runtime API
### Add Runtime

#### RPC Message 
Needs to be signed by a valid Colony Private Key.

```json
{
    "method": "addruntime",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
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
        "status": 0
    }
}
```

#### Decoded reply

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
    "status": 0
}
```

### List Runtimes

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "getruntimes",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "colonyid": "863e313bfd882fe7c0f13c14aff1f3f02ba763bcb48377e50d505289c81e47b6"
}
```

#### Decoded reply

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
        "status": 1
    }
]
```

###  Get Runtime info

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "getruntime",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "runtimeid": "ed2aa78eabe3d1f6fd46ef1247199e9a12faf1a8f1bcba0db51265515c3f08e0"
}
```

#### Decoded reply

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
    "status": 2
}
```

### Approve Runtime 

#### RPC Message 
Needs to be signed by a valid Colony Private Key.

```json
{
    "method": "approveruntime",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded Payload

```json
{
    "runtimeid": "e40e2862e3a68e1c79af4e9475ef64fbf588e13619f4daa7183673b34e189c87"
}
```

#### Decoded Reply
None

###  Reject Runtime 

#### RPC Message 
Needs to be signed by a valid Colony Private Key.

```json
{
    "method": "rejectruntime",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "runtimeid": "7804cea6a50f2a258ad815b0ed37b6b312c813bf7387cef04958971335faae21"
}
```

#### Decoded reply
None

## Process API

### Submit Process Specification 

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "submitprocessspec",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
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

#### Decoded reply

```json
{
    "processid": "2c0fd0407292538cb8dce3cb306f88b2ab7f3726d649e07502eb04344d9f7164",
    "assignedruntimeid": "",
    "isassigned": false,
    "status": 0,
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

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "assignprocess",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Encoded payload

```json
{
    "colonyid": "326691e2b5fc0651b5d781393c7279ab3dc58c6627d0a7b2a09e9aa0e4a60950"
}
```

#### Decoded reply

```json
{
    "processid": "68db01b27271168cb1011c1c54cc31a54f23eb7e5767e49bb34fb206591d2a65",
    "assignedruntimeid": "d02274979e69d534202ca4cdcb3847c56e860d09039399feee6358b8c285d502",
    "isassigned": true,
    "status": 1,
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

###  List processes

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "getprocesses",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

```json
{
    "coloyid": "891f0c88e8a00cb103df472e4ece347a41eb0115e5c40f12d565bb24eb3fc71d",
    "count": 2,
    "state": 3 
}
```

#### Decoded reply 

```json
[
    {
        "processid": "88169d23b0828ed65f0a007e4be6bf9734358b9a64379d0c6e53a0496216db4c",
        "assignedruntimeid": "653c818113e878d704935e639371f72a3167d510008607c70176e8147adf7865",
        "isassigned": true,
        "status": 3,
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

###  Get Process info

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "getprocess",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload 

```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce"
}
```

#### Decoded reply

```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedruntimeid": "",
    "isassigned": false,
    "status": 0,
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

### Mark Process as Successful 

#### RPC Message 
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

```json
{
    "method": "marksuccessful",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "processid": "ed041355071d2ee6d0ec27b480e2e4c8006cf465ec408b57fcdaa5dac76af8e2"
}
```

#### Decoded reply
None

### Mark a Proceess as Failed 

#### RPC Message 
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

```json
{
    "method": "markfailed",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "processid": "24f6d85804e2abde0c85a9e8aef8b308c44a72323565b14f11756d4997acf200"
}
```

#### Decoded reply
None

###  Add Attribute to a Process 

#### RPC Message 
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

```json
{
    "method": "addattribute",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload

```json
{
    "attribute": {
        "attributeid": "216e26cb089032d2f941454e7db5f3ae1591eeb43eb477c3f8ed545b96d4f690",
        "targetid": "c4775cab695da8a77b503bbe29df8ae39dafd1c7fed3275dac11b436c1724dbf",
        "attributetype": 1,
        "key": "result",
        "value": "helloworld"
    }
}
```

#### Decoded reply

```json
{
    "attributeid": "216e26cb089032d2f941454e7db5f3ae1591eeb43eb477c3f8ed545b96d4f690",
    "targetid": "c4775cab695da8a77b503bbe29df8ae39dafd1c7fed3275dac11b436c1724dbf",
    "attributetype": 1,
    "key": "result",
    "value": "helloworld"
}
```

###  Get Attribute assigned to a Process 

#### RPC Message 
Needs to be signed by a valid Runtime Private Key.

```json
{
    "method": "gettattribute",
    "signature": "...",
    "payload": "...",
    "error": false
}
```

#### Decoded payload 

```json
{
    "attributeid": "a1d8f3613e074a250c2fbab478a0e11eb40defee66bd9b6a6ceb96990f1486eb"
}
```

#### Decoded reply

```json
{
    "attributeid": "a1d8f3613e074a250c2fbab478a0e11eb40defee66bd9b6a6ceb96990f1486eb",
    "targetid": "3d893a44a30c7e5c5c595413a9de1545a9d43a844528831c4e205b280c074e56",
    "attributetype": 1,
    "key": "result",
    "value": "helloworld"
}
```
