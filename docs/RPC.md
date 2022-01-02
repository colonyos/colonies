# HTTP RPC protocol
* Messages are POSTed to http://SERVERHOST:SERVERPORT/api 
* It is expected that the signature is set as a HTTP Header named **Signature**.
* If the HTTP status code is not 200 OK, Reply messages can also contain an error message formatted as below.
```json
{
    "stats": "500",
    "name": "something when wrong here"
}
```

## Colony API
### Add Colony
Needs to be signed by a valid Server Owner Private Key.

#### Message

```json
{
    "rpc": {
        "method": "addcolony",
        "nonce": "5681b8c0e9f966df9b51e37e351449ad50a315baf20023ce0d24666dad59b991"
    },
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
Needs to be signed by a valid Server Owner Private Key.
#### Message
```json
{
    "rpc": {
        "method": "getcolonies",
        "nonce": "723c2f48b5654cf420cb259ef6e64d8f9168ea4a0ffd6d221f7e950d2d3c567c"
    }
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
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "getcolony",
        "nonce": "ed8b9825e4ad9e0864b4a1d363f387aa5c9ff2e67db4c7eb8174246a83e66e44"
    },
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
### Add Runtime
Needs to be signed by a valid Colony Private Key.

#### Message
```json
{
    "rpc": {
        "method": "addruntime",
        "nonce": "4431c27e78c5bee01c92ffc176747af35a75662a66be8f2065eaf6ba41befc9d"
    },
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
    "status": 0
}
```

### List Runtimes
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "getruntimes",
        "nonce": "70ecd5e900fff8a8360062530e2a30ee9d4b759e5bf8b99e6ae486ea02c0c42a"
    },
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
        "status": 1
    }
]
```

###  Get Runtime info
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "getruntime",
        "nonce": "3c70ab9a74ef7f3bfe13136e9ee90d35cba67458b165a0e7cb0384fee5c41312"
    },
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
    "status": 2
}
```

### Approve Runtime 
Needs to be signed by a valid Colony Private Key.

#### Message
```json
{
    "rpc": {
        "method": "approveruntime",
        "nonce": "7065c7e7092b32a303c48e7fd0267bff71ac90a0dd165af971097c3ae94b4688"
    },
    "runtimeid": "e40e2862e3a68e1c79af4e9475ef64fbf588e13619f4daa7183673b34e189c87"
}
```

#### Reply 
None

###  Reject Runtime 
Needs to be signed by a valid Colony Private Key.

#### Message
```json
{
    "rpc": {
        "method": "rejectruntime",
        "nonce": "c78a97436c2a8a37438f0d811882248ed0700b54ca46acd7791fa2f40c8f02ee"
    },
    "runtimeid": "7804cea6a50f2a258ad815b0ed37b6b312c813bf7387cef04958971335faae21"
}
```

#### Reply 
None

## Process API

### Submit Process Specification 
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "submitprocessspec",
        "nonce": "ea23df61613c540b05807b9fdccacbb05c80e62776bc640b46307eaef6b3bcde"
    },
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
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "assignprocessspec",
        "nonce": "bcc4051c250db9f14d30c3ddba2e9eefded526ee49f255e4d1c1c9c0761dc145"
    },
    "colonyid": "326691e2b5fc0651b5d781393c7279ab3dc58c6627d0a7b2a09e9aa0e4a60950"
}
```

#### Reply 
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
Needs to be signed by a valid Runtime Private Key.

The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

#### Message
```json
{
    "rpc": {
        "method": "getprocesses",
        "nonce": "7ad27ab65779ceee8b6796489c3b349ecd51c68e44ed64ed37f3d2fde129d85e"
    },
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
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "getprocess",
        "nonce": "0eed2d1bf222767d2b1c4f3807f0a38f775853d480fd5e3d64ddcd9d288f95d3"
    },
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce"
}
```

#### Reply 
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
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

#### Message
```json
{
    "rpc": {
        "method": "marksuccessful",
        "nonce": "a9e695209156eb0f63c1077afc3e42c91d4b538abca4e654f6cfb7895390658e"
    },
    "processid": "ed041355071d2ee6d0ec27b480e2e4c8006cf465ec408b57fcdaa5dac76af8e2"
}
```
#### Reply 
None

### Mark a Proceess as Failed 
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

#### Message
```json
{
    "rpc": {
        "method": "markfailed",
        "nonce": "963058bab6cea72ddfe6b7ba6d265e9f6b4837ed4d937cae126e4084c3d0e4cf"
    },
    "processid": "24f6d85804e2abde0c85a9e8aef8b308c44a72323565b14f11756d4997acf200"
}
```
#### Reply 
None

###  Add Attribute to a Process 
Needs to be signed by a valid Runtime Private Key. The Runtime ID needs to match the RuntimeID assigned to the process.

#### Message
```json
{
    "rpc": {
        "method": "addattribute",
        "nonce": "0e1f50e74d171217e77cb0fcfd54656b65aba48c57fad055966b523ebc4196ed"
    },
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

###  Get Attribute assigned to a Process 
Needs to be signed by a valid Runtime Private Key.

#### Message
```json
{
    "rpc": {
        "method": "getattribute",
        "nonce": "6e36ae66c79467899e88270e5854eb6fe15f3595446aebbf236620617e66fc30"
    },
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
