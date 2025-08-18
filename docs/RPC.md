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
    "status": "500",
    "message": "something when wrong here"
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
* PayloadType: [**addcolonymsg**](../pkg/rpc/add_colony_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: A [Colony object](../pkg/core/colony.go)

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

### Remove Colony
* PayloadType: [**removecolonymsg**](../pkg/rpc/remove_colony_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removecolonymsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
{}
```

### List Colonies
* PayloadType: [**getcoloniesmsg**](../pkg/rpc/get_colonies_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: An array of [Colony objects](../pkg/core/colony.go)

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
* PayloadType: [**getcolonymsg**](../pkg/rpc/get_colony_msg.go)
* Credentials: A valid Exectutor Private Key
* Returns: A [Colony object](../pkg/core/colony.go)

#### Payload 
```json
{
    "msgtype": "getcolonymsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
{
    "colonyid": "ac8dc8949af395fd51ead31d598b252bda02f1f6eed11ace7fc7dc8dd85ac32e",
    "name": "test_colony_name"
}
```

## Executor API

### Add Executor
* PayloadType: [**addexecutormsg**](../pkg/rpc/add_executor_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An [Executor object](../pkg/core/executor.go)

*Note: The state and commissiontime fields are set by the server. The executor is always created in a PENDING state.*

#### Payload 
```json
{
    "msgtype": "addexecutormsg",
    "executor": {
        "executorid": "38df5bbbcf0ccb438d2e4151638e3967bf28a5654af6a7e5acc590c0e49fae06",
        "executortype": "test_executor_type",
        "name": "test_executor_name",
        "colonyname": "my_colony_name",
        "capabilities": {
            "hardware": {
                "cpu": "AMD Ryzen 9 5950X",
                "mem": "80326MB",
                "nodes": 1
            }
        }
    }
}
```

#### Reply 
```json
{
    "executorid": "38df5bbbcf0ccb438d2e4151638e3967bf28a5654af6a7e5acc590c0e49fae06",
    "executortype": "test_executor_type",
    "executorname": "test_executor_name",
    "colonyname": "my_colony_name",
    "state": 0,
    "requirefuncreg": false,
    "commissiontime": "2022-01-02T12:00:00Z",
    "lastheardfromtime": "2022-01-02T12:05:00Z",
    "location": {
        "long": 0.0,
        "lat": 0.0,
        "desc": "test location"
    },
    "capabilities": {
        "hardware": {
            "model": "test_model",
            "nodes": 1,
            "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
            "mem": "80326MB",
            "storage": "1TB",
            "gpu": {
                "name": "NVIDIA GeForce RTX 2080 Ti Rev. A",
                "mem": "11GB",
                "count": 1,
                "nodecount": 1
            }
        },
        "software": {
            "name": "test_software",
            "type": "container",
            "version": "1.0.0"
        }
    },
    "allocations": {
        "projects": {}
    }
}
```

### List Executors
* PayloadType: [**getexecutorsmsg**](../pkg/rpc/get_executors_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Executor objects](../pkg/core/executor.go)

#### Payload 
```json
{
    "msgtype": "getexecutorsmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
[
    {
        "executorid": "9525365b67efdbbf37bc1fa7628c7e75bafd2f298cd26f75500bc1364b2c4c1c",
        "executortype": "test_executor_type",
        "executorname": "test_executor_name",
        "colonyname": "my_colony_name",
        "state": 1,
        "requirefuncreg": false,
        "commissiontime": "2022-01-02T12:00:00Z",
        "lastheardfromtime": "2022-01-02T12:05:00Z",
        "location": {
            "long": 0.0,
            "lat": 0.0,
            "desc": "test location"
        },
        "capabilities": {
            "hardware": {
                "model": "test_model",
                "nodes": 1,
                "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
                "mem": "80326MB",
                "storage": "1TB",
                "gpu": {
                    "name": "NVIDIA GeForce RTX 2080 Ti Rev. A",
                    "mem": "11GB",
                    "count": 1,
                    "nodecount": 1
                }
            },
            "software": {
                "name": "test_software",
                "type": "container",
                "version": "1.0.0"
            }
        },
        "allocations": {
            "projects": {}
        }
    }
]
```

### Get Executor info
* PayloadType: [**getexecutormsg**](../pkg/rpc/get_executor_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An [Executor object](../pkg/core/executor.go)

#### Payload 
```json
{
    "msgtype": "getexecutormsg",
    "colonyname": "my_colony_name",
    "executorname": "my_executor_name"
}
```

#### Reply 
```json
{
    "executorid": "ed2aa78eabe3d1f6fd46ef1247199e9a12faf1a8f1bcba0db51265515c3f08e0",
    "executortype": "test_executor_type",
    "executorname": "test_executor_name",
    "colonyname": "my_colony_name",
    "state": 2,
    "requirefuncreg": false,
    "commissiontime": "2022-01-02T12:00:00Z",
    "lastheardfromtime": "2022-01-02T12:05:00Z",
    "location": {
        "long": 0.0,
        "lat": 0.0,
        "desc": "test location"
    },
    "capabilities": {
        "hardware": {
            "model": "test_model",
            "nodes": 1,
            "cpu": "AMD Ryzen 9 5950X (32) @ 3.400GHz",
            "mem": "80326MB",
            "storage": "1TB",
            "gpu": {
                "name": "NVIDIA GeForce RTX 2080 Ti Rev. A",
                "mem": "11GB",
                "count": 1,
                "nodecount": 1
            }
        },
        "software": {
            "name": "test_software",
            "type": "container",
            "version": "1.0.0"
        }
    },
    "allocations": {
        "projects": {}
    }
}
```

### Approve Executor 
* PayloadType: [**approveexecutormsg**](../pkg/rpc/approve_executor_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload
```json
{
    "msgtype": "approveexecutormsg",
    "colonyname": "my_colony_name",
    "executorname": "my_executor_name"
}
```

#### Reply
```json
{}
```

### Reject Executor 
* PayloadType: [**rejectexecutormsg**](../pkg/rpc/reject_executor_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "rejectexecutormsg",
    "colonyname": "my_colony_name",
    "executorname": "my_executor_name"
}
```

#### Reply 
```json
{}
```

### Remove Executor 
* PayloadType: [**removeexecutormsg**](../pkg/rpc/remove_executor_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeexecutormsg",
    "colonyname": "my_colony_name",
    "executorname": "my_executor_name"
}
```

#### Reply 
```json
{}
```

### Report Allocations
* PayloadType: [**reportallocationmsg**](../pkg/rpc/report_allocation_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "reportallocationmsg",
    "colonyname": "my_colony_name",
    "executorname": "my_executor_name",
    "allocations": {
        "projects": {
            "project-123": {
                "projectid": "project-123",
                "cpu": "2000m",
                "mem": "4Gi"
            }
        }
    }
}
```

#### Reply 
```json
{}
```

## Process API

### Submit Process Specification 
* PayloadType: [**submitfuncspecmsg**](../pkg/rpc/submit_funcspec_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Process object](../pkg/core/process.go)

#### Payload 
```json
{
    "msgtype": "submitfuncspecmsg",
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyid": "2de470e10b87dc261c05f6b2da45d0802044208d6c617a056f4824d958710827",
            "executornames": [],
            "executortype": "test_executor_type",
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
    "assignedexecutorid": "",
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
            "executornames": [],
            "executortype": "test_executor_type",
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

### Assign Process to a Executor 
* PayloadType: [**assignprocessmsg**](../pkg/rpc/assign_process_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Process object](../pkg/core/process.go)

#### Payload 
```json
{
    "msgtype": "assignprocessmsg",
    "colonyname": "my_colony_name",
    "timeout": -1,
    "availablecpu": "1000m",
    "availablemem": "1000Mi"
}
```

#### Reply 
```json
{
    "processid": "68db01b27271168cb1011c1c54cc31a54f23eb7e5767e49bb34fb206591d2a65",
    "assignedexecutorid": "d02274979e69d534202ca4cdcb3847c56e860d09039399feee6358b8c285d502",
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
            "executornames": [],
            "executortype": "test_executor_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### List process history
* PayloadType: [**getprocesshistmsg**](../pkg/rpc/get_process_hist.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Process objects](../pkg/core/process.go)

#### Payload 
The state attribute can have the following values:
* 0 : Waiting 
* 1 : Running 
* 2 : Success 
* 3 : Failed 

Note, all process will be returned for the entire colony if executorID is not specified.

```json
{
    "msgtype": "getprocesshistmsg",
    "colonyname": "891f0c88e8a00cb103df472e4ece347a41eb0115e5c40f12d565bb24eb3fc71d",
    "executorid": "",
    "seconds": 100,
    "state": 3 
}
```

#### Reply 
```json
[
    {
        "processid": "88169d23b0828ed65f0a007e4be6bf9734358b9a64379d0c6e53a0496216db4c",
        "assignedexecutorid": "653c818113e878d704935e639371f72a3167d510008607c70176e8147adf7865",
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
                "executornames": [],
                "executortype": "test_executor_type",
                "mem": 1000,
                "cores": 10,
                "gpus": 1
            },
            "env": {}
        }
    }
]
```

### List processes
* PayloadType: [**getprocessesmsg**](../pkg/rpc/get_processes_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Process objects](../pkg/core/process.go)

#### Payload 
The state attribute can have the following values:
* 0 : Waiting 
* 1 : Running 
* 2 : Success 
* 3 : Failed 

```json
{
    "msgtype": "getprocessesmsg",
    "colonyname": "891f0c88e8a00cb103df472e4ece347a41eb0115e5c40f12d565bb24eb3fc71d",
    "count": 2,
    "state": 3,
    "executortype": "test_executor_type",
    "label": "",
    "initiator": ""
}
```

#### Reply 
```json
[
    {
        "processid": "88169d23b0828ed65f0a007e4be6bf9734358b9a64379d0c6e53a0496216db4c",
        "assignedexecutorid": "653c818113e878d704935e639371f72a3167d510008607c70176e8147adf7865",
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
                "executornames": [],
                "executortype": "test_executor_type",
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
* PayloadType: [**getprocessmsg**](../pkg/rpc/get_process_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Process object](../pkg/core/process.go)

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
    "assignedexecutorid": "",
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
            "executornames": [],
            "executortype": "test_executor_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### Remove Process
* PayloadType: [**removeprocessmsg**](../pkg/rpc/remove_process_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeprocessmsg",
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce"
}
```

#### Reply 
```json
{}
```

### Remove all Processes
* PayloadType: [**removeallprocessesmsg**](../pkg/rpc/remove_all_processes_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeallprocessesmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
{}
```

### Close Process as Successful 
* PayloadType: [**closesuccessfulmsg**](../pkg/rpc/close_successful_msg.go)
* Credentials: A valid Executor Private Key and the Executor ID needs to match the ExecutorID assigned to the process
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "closesuccessfulmsg",
    "processid": "ed041355071d2ee6d0ec27b480e2e4c8006cf465ec408b57fcdaa5dac76af8e2"
    "out": []
}
```

#### Reply
```json
{}
```

### Close a Process as Failed 
* PayloadType: [**closefailedmsg**](../pkg/rpc/close_failed_msg.go)
* Credentials: A valid Executor Private Key and the Executor ID needs to match the ExecutorID assigned to the process
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "closefailedmsg",
    "processid": "24f6d85804e2abde0c85a9e8aef8b308c44a72323565b14f11756d4997acf200"
    "errors": []
}
```

#### Reply 
```json
{}
```

### Get Colony Statistics 
* PayloadType: [**getcolonystatsmsg**](../pkg/rpc/get_colony_statistics_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: A [Statistics object](../pkg/core/statistics.go)

#### Payload 
```json
{
    "msgtype": "getcolonystatsmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
{
    "colonies": 1,
    "executors": 5,
    "waitingprocesses": 1,
    "runningprocesses": 2,
    "successfulprocesses": 3,
    "failedprocesses": 4,
    "waitingworkflows": 0,
    "runningworkflows": 0,
    "successfulworkflows": 0,
    "failedworkflows": 0
}
```


### Add Attribute to a Process 
* PayloadType: [**addattributemsg**](../pkg/rpc/add_attribute_msg.go)
* Credentials: A valid Executor Private Key and the Executor ID needs to match the ExecutorID assigned to the process
* Returns: An [Attribute object](../pkg/core/attribute.go)

*Note: The attributeid and targetprocessgraphid fields are generated by the server and will be ignored if provided.*

#### Payload 
```json
{
    "msgtype": "addattributemsg",
    "attribute": {
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
* PayloadType: [**getattributemsg**](../pkg/rpc/get_attribute_msg.go)
* Credentials: A valid Executor  Private Key
* Returns: An [Attribute object](../pkg/core/attribute.go)

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
* PayloadType: [**subscribeprocessmsg**](../pkg/rpc/subscribe_process_msg.go)
* Credentials: A valid Executor Private Key
* Comments: Receives an event when a process changes state. The payload needs to be sent over a websocket to: wss://host:port/pubsub
* Returns: A [Process object](../pkg/core/process.go)

#### Payload 
The state attribute can have the following values:
* 0 : Waiting 
* 1 : Running 
* 2 : Success 
* 3 : Failed 

```json
{
    "msgtype": "subscribeprocessmsg",
    "colonyname": "my_colony_name",
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "executortype": "test_executor_type",
    "state": 1,
    "timeout": -1
}
```

#### Reply 
```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedexecutorid": "",
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
            "executornames": [],
            "executortype": "test_executor_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

### Set Process Output
* PayloadType: [**setoutputmsg**](../pkg/rpc/set_output_msg.go)
* Credentials: A valid Executor Private Key and the Executor ID needs to match the ExecutorID assigned to the process
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "setoutputmsg",
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "out": ["result1", "result2", {"key": "value"}]
}
```

#### Reply 
```json
{}
```

### Subscribe Processes Events
* PayloadType: [**subscribeprocessesmsg**](../pkg/rpc/subscribe_processes_msg.go)
* Credentials: A valid Executor Private Key
* Comments: Receives an event when processes are added or change state. The payload needs to be sent over a websocket to: wss://host:port/pubsub
* Returns: A [Process object](../pkg/core/process.go)

#### Payload 
The state attribute can have the following values:
* 1 : Waiting 
* 2 : Running 
* 3 : Success 
* 4 : Failed 

```json
{
    "msgtype": "subscribeprocessesmsg",
    "colonyname": "my_colony_name",
    "executortype": "test_executor_type",
    "state": 1,
    "timeout": -1
}
```

#### Reply 
```json
{
    "processid": "80a98f46c7a364fd33339a6fb2e6c5d8988384fdbf237b4012490c4658bbc9ce",
    "assignedexecutorid": "",
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
            "executornames": [],
            "executortype": "test_executor_type",
            "mem": 1000,
            "cores": 10,
            "gpus": 1
        },
        "env": {}
    }
}
```

## Workflow & Process Graph API

### Submit Workflow Specification 
* PayloadType: [**submitworkflowspecmsg**](../pkg/rpc/submit_workflow_spec.go)
* Credentials: A valid Executor Private Key
* Returns: A [ProcessGraph object](../pkg/core/processgraph.go)

#### Payload 
```json
{
    "msgtype": "submitworkflowspecmsg",
    "spec": {
        "name": "my_workflow",
        "colonyname": "my_colony_name",
        "funcspecs": [
            {
                "timeout": -1,
                "maxretries": 3,
                "conditions": {
                    "colonyname": "my_colony_name",
                    "executortype": "test_executor_type",
                    "mem": 1000,
                    "cores": 10,
                    "gpus": 1
                },
                "env": {
                    "test_key": "test_value"
                }
            }
        ]
    }
}
```

#### Reply 
```json
{
    "processgraphid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "processids": ["process-id-1", "process-id-2"],
    "state": 0,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z"
}
```

### Add Child to Process Graph
* PayloadType: [**addchildmsg**](../pkg/rpc/add_child_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Process object](../pkg/core/process.go)

#### Payload 
```json
{
    "msgtype": "addchildmsg",
    "processgraphid": "a-valid-sha256-hash-id",
    "parentprocessid": "parent-process-id",
    "childprocessid": "child-process-id",
    "insert": false,
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyname": "my_colony_name",
            "executortype": "test_executor_type"
        },
        "env": {
            "child_key": "child_value"
        }
    }
}
```

#### Reply
```json
{
    "processid": "new-child-process-id",
    "assignedexecutorid": "",
    "isassigned": false,
    "state": 0,
    "submissiontime": "2022-01-02T12:10:00.000000Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z",
    "deadline": "0001-01-01T00:00:00Z",
    "retries": 0,
    "attributes": [],
    "spec": {
        "timeout": -1,
        "maxretries": 3,
        "conditions": {
            "colonyname": "my_colony_name",
            "executortype": "test_executor_type"
        },
        "env": {
            "child_key": "child_value"
        }
    }
}
```

### Get Process Graph
* PayloadType: [**getprocessgraphmsg**](../pkg/rpc/get_processgraph.go)
* Credentials: A valid Executor Private Key
* Returns: A [ProcessGraph object](../pkg/core/processgraph.go)

#### Payload 
```json
{
    "msgtype": "getprocessgraphmsg",
    "processgraphid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{
    "processgraphid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "processids": ["process-id-1", "process-id-2"],
    "state": 2,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "2022-01-02T12:08:20.000000Z",
    "endtime": "2022-01-02T12:09:30.000000Z"
}
```

### List Process Graphs
* PayloadType: [**getprocessgraphsmsg**](../pkg/rpc/get_processgraphs.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [ProcessGraph objects](../pkg/core/processgraph.go)

#### Payload 
```json
{
    "msgtype": "getprocessgraphsmsg",
    "colonyname": "my_colony_name",
    "count": 10,
    "state": 2
}
```

#### Reply 
```json
[
    {
        "processgraphid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "processids": ["process-id-1", "process-id-2"],
        "state": 2,
        "submissiontime": "2022-01-02T12:08:16.226133Z",
        "starttime": "2022-01-02T12:08:20.000000Z",
        "endtime": "2022-01-02T12:09:30.000000Z"
    }
]
```

### Remove Process Graph
* PayloadType: [**removeprocessgraphmsg**](../pkg/rpc/remove_processgraph_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeprocessgraphmsg",
    "processgraphid": "a-valid-sha256-hash-id",
    "all": false
}
```

#### Reply 
```json
{}
```

### Remove All Process Graphs
* PayloadType: [**removeallprocessgraphsmsg**](../pkg/rpc/remove_all_processgraphs_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeallprocessgraphsmsg",
    "colonyname": "my_colony_name",
    "state": 2
}
```

#### Reply 
```json
{}
```

## Cron API

### Add Cron
* PayloadType: [**addcronmsg**](../pkg/rpc/add_cron_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Cron object](../pkg/core/cron.go)

*Note: The cronid, initiatorid, initiatorname, and all state fields (nextrun, lastrun, prevprocessgraphid) are managed by the server and will be ignored if provided.*

#### Payload 
```json
{
    "msgtype": "addcronmsg",
    "cron": {
        "name": "my_cron_job",
        "colonyname": "my_colony_name",
        "cronexpression": "0 0 * * *",
        "interval": -1,
        "random": false,
        "waitforprevprocessgraph": true,
        "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\"}}]}"
    }
}
```

#### Reply 
```json
{
    "cronid": "a-valid-sha256-hash-id",
    "name": "my_cron_job",
    "colonyname": "my_colony_name",
    "cronexpression": "0 0 * * *",
    "interval": 86400,
    "random": false,
    "nextrun": "2022-01-03T00:00:00Z",
    "lastrun": "0001-01-01T00:00:00Z",
    "prevprocessgraphid": "",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}"
}
```

### Get Cron
* PayloadType: [**getcronmsg**](../pkg/rpc/get_cron_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Cron object](../pkg/core/cron.go)

#### Payload 
```json
{
    "msgtype": "getcronmsg",
    "cronid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{
    "cronid": "a-valid-sha256-hash-id",
    "name": "my_cron_job",
    "colonyname": "my_colony_name",
    "cronexpression": "0 0 * * *",
    "interval": 86400,
    "random": false,
    "nextrun": "2022-01-03T00:00:00Z",
    "lastrun": "2022-01-02T00:00:00Z",
    "prevprocessgraphid": "prev-process-graph-id",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}"
}
```

### List Crons
* PayloadType: [**getcronsmsg**](../pkg/rpc/get_crons_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Cron objects](../pkg/core/cron.go)

#### Payload 
```json
{
    "msgtype": "getcronsmsg",
    "colonyname": "my_colony_name",
    "count": 10
}
```

#### Reply 
```json
[
    {
        "cronid": "a-valid-sha256-hash-id",
        "name": "my_cron_job",
        "colonyname": "my_colony_name",
        "cronexpression": "0 0 * * *",
        "interval": 86400,
        "random": false,
        "nextrun": "2022-01-03T00:00:00Z",
        "lastrun": "2022-01-02T00:00:00Z",
        "prevprocessgraphid": "prev-process-graph-id",
        "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}"
    }
]
```

### Remove Cron
* PayloadType: [**removecronmsg**](../pkg/rpc/remove_cron_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removecronmsg",
    "cronid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{}
```

### Run Cron
* PayloadType: [**runcronmsg**](../pkg/rpc/run_cron_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Cron object](../pkg/core/cron.go)

#### Payload 
```json
{
    "msgtype": "runcronmsg",
    "cronid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{
    "cronid": "a-valid-sha256-hash-id",
    "initiatorid": "initiator-id",
    "initiatorname": "initiator_name",
    "colonyname": "my_colony_name",
    "name": "my_cron_job",
    "cronexpression": "0 0 * * *",
    "interval": 86400,
    "random": false,
    "nextrun": "2022-01-03T00:00:00Z",
    "lastrun": "2022-01-02T00:00:00Z",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}",
    "prevprocessgraphid": "prev-process-graph-id",
    "waitforprevprocessgraph": false,
    "checkerperiod": 60
}
```

## Generator API

### Add Generator
* PayloadType: [**addgeneratormsg**](../pkg/rpc/add_generator_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Generator object](../pkg/core/generator.go)

*Note: The generatorid, initiatorid, initiatorname, and all state fields (lastrun, counter) are managed by the server and will be ignored if provided.*

#### Payload 
```json
{
    "msgtype": "addgeneratormsg",
    "generator": {
        "name": "my_generator",
        "colonyname": "my_colony_name",
        "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\"}}]}",
        "trigger": 5,
        "timeout": 60
    }
}
```

#### Reply 
```json
{
    "generatorid": "a-valid-sha256-hash-id",
    "name": "my_generator",
    "colonyname": "my_colony_name",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}",
    "trigger": 5,
    "counter": 0,
    "lastrun": "0001-01-01T00:00:00Z"
}
```

### Get Generator
* PayloadType: [**getgeneratormsg**](../pkg/rpc/get_generator_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Generator object](../pkg/core/generator.go)

#### Payload 
```json
{
    "msgtype": "getgeneratormsg",
    "generatorid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{
    "generatorid": "a-valid-sha256-hash-id",
    "name": "my_generator",
    "colonyname": "my_colony_name",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}",
    "trigger": 5,
    "counter": 3,
    "lastrun": "2022-01-02T12:00:00Z"
}
```

### Resolve Generator by Name
* PayloadType: [**resolvegeneratormsg**](../pkg/rpc/resolve_generator_msg.go)
* Credentials: A valid Executor or User Private Key
* Returns: A [Generator object](../pkg/core/generator.go)

#### Payload 
```json
{
    "msgtype": "resolvegeneratormsg",
    "colonyname": "my_colony_name",
    "generatorname": "my_generator"
}
```

#### Reply 
```json
{
    "generatorid": "a-valid-sha256-hash-id",
    "name": "my_generator",
    "colonyname": "my_colony_name",
    "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}",
    "trigger": 5,
    "counter": 3,
    "lastrun": "2022-01-02T12:00:00Z"
}
```

### List Generators
* PayloadType: [**getgeneratorsmsg**](../pkg/rpc/get_generators_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Generator objects](../pkg/core/generator.go)

#### Payload 
```json
{
    "msgtype": "getgeneratorsmsg",
    "colonyname": "my_colony_name",
    "count": 10
}
```

#### Reply 
```json
[
    {
        "generatorid": "a-valid-sha256-hash-id",
        "name": "my_generator",
        "colonyname": "my_colony_name",
        "workflowspec": "{\"name\":\"my_workflow\",\"colonyname\":\"my_colony_name\",\"funcspecs\":[{\"timeout\":-1,\"maxretries\":3,\"conditions\":{\"colonyname\":\"my_colony_name\",\"executortype\":\"test_executor_type\",\"mem\":1000,\"cores\":10,\"gpus\":1},\"env\":{\"test_key\":\"test_value\"}}]}",
        "trigger": 5,
        "counter": 3,
        "lastrun": "2022-01-02T12:00:00Z"
    }
]
```

### Remove Generator
* PayloadType: [**removegeneratormsg**](../pkg/rpc/remove_generator_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removegeneratormsg",
    "generatorid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{}
```

### Pack Generator
* PayloadType: [**packgeneratormsg**](../pkg/rpc/pack_generator.go)
* Credentials: A valid Executor Private Key
* Returns: A [ProcessGraph object](../pkg/core/processgraph.go)

#### Payload 
```json
{
    "msgtype": "packgeneratormsg",
    "generatorid": "a-valid-sha256-hash-id",
    "arg": "data-payload-to-add"
}
```

#### Reply 
```json
{
    "processgraphid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "processids": ["process-id-1", "process-id-2"],
    "state": 0,
    "submissiontime": "2022-01-02T12:08:16.226133Z",
    "starttime": "0001-01-01T00:00:00Z",
    "endtime": "0001-01-01T00:00:00Z"
}
```

## File Management API

### Add File
* PayloadType: [**addfilemsg**](../pkg/rpc/add_file_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [File object](../pkg/core/file.go)

*Note: The fileid and added timestamp are set by the server and will be ignored if provided.*

#### Payload 
```json
{
    "msgtype": "addfilemsg",
    "file": {
        "colonyname": "my_colony_name",
        "label": "my_file_label",
        "name": "example.txt",
        "size": 1024,
        "checksum": "checksum-hash",
        "checksumalg": "SHA256"
    }
}
```

#### Reply 
```json
{
    "fileid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "label": "my_file_label",
    "name": "example.txt",
    "size": 1024,
    "checksum": "checksum-hash",
    "checksumtype": "SHA256",
    "added": "2022-01-02T12:00:00Z"
}
```

### Get File
* PayloadType: [**getfilemsg**](../pkg/rpc/get_file_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An array of [File objects](../pkg/core/file.go)

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "getfilemsg",
    "colonyname": "my_colony_name",
    // Use one of the following methods to identify the file(s).
    // Method 1: Get a specific file by its unique ID (highest priority).
    "fileid": "a-valid-sha256-hash-id",

    // Method 2: Get files by label and name (used if fileid is empty).
    "label": "my_file_label",
    "name": "example.txt",
    "latest": true // Set to true to get only the most recent version.
}
```

#### Reply 
```json
[
    {
        "fileid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "label": "my_file_label",
        "name": "example.txt",
        "size": 1024,
        "checksum": "checksum-hash",
        "checksumtype": "SHA256",
        "added": "2022-01-02T12:00:00Z"
    }
]
```

### List Files
* PayloadType: [**getfilesmsg**](../pkg/rpc/get_files_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [FileData objects](../pkg/core/filedata.go)

#### Payload 
```json
{
    "msgtype": "getfilesmsg",
    "colonyname": "my_colony_name",
    "label": "my_file_label"
}
```

#### Reply 
```json
[
    {
        "name": "example.txt",
        "checksum": "checksum-hash",
        "size": 1024,
        "s3filename": "s3-object-filename"
    }
]
```

### List File Labels
* PayloadType: [**getfilelabelsmsg**](../pkg/rpc/get_filelabels_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Label objects](../pkg/core/label.go)

#### Payload 
```json
{
    "msgtype": "getfilelabelsmsg",
    "colonyname": "my_colony_name",
    "name": "data_files",
    "exact": false
}
```

#### Reply 
```json
[
    {
        "name": "my_file_label",
        "files": 5
    },
    {
        "name": "another_label", 
        "files": 3
    },
    {
        "name": "data_files",
        "files": 12
    }
]
```

### Remove File
* PayloadType: [**removefilemsg**](../pkg/rpc/remove_file_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "removefilemsg",
    "colonyname": "my_colony_name",
    // Use one of the following methods to identify the file(s) for deletion.
    // Method 1: Delete a specific file by its unique ID (highest priority).
    "fileid": "a-valid-sha256-hash-id",

    // Method 2: Delete all versions of a file by label and name (used if fileid is empty).
    "label": "my_file_label",
    "name": "example.txt"
}
```

#### Reply 
```json
{}
```

## Logging API

### Add Log
* PayloadType: [**addlogmsg**](../pkg/rpc/add_log_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "addlogmsg",
    "processid": "a-valid-sha256-hash-id",
    "message": "This is a log message"
}
```

#### Reply 
```json
{}
```

### Get Logs
* PayloadType: [**getlogsmsg**](../pkg/rpc/get_logs_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Log objects](../pkg/core/log.go)

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "getlogsmsg",
    "colonyname": "my_colony_name",
    // Specify either executorname or processid.
    // If executorname is provided, processid is ignored.
    "processid": "a-valid-sha256-hash-id",
    "executorname": "my_executor_name",
    "count": 100,
    "since": 1640995200
}
```

#### Reply 
```json
[
    {
        "processid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "executorname": "my_executor_name",
        "message": "Log message 1",
        "timestamp": 1640995200
    },
    {
        "processid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "executorname": "my_executor_name",
        "message": "Log message 2",
        "timestamp": 1640995201
    }
]
```

### Search Logs
* PayloadType: [**searchlogsmsg**](../pkg/rpc/search_logs_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Log objects](../pkg/core/log.go)

#### Payload 
```json
{
    "msgtype": "searchlogsmsg",
    "colonyname": "my_colony_name",
    "text": "error",
    "count": 50,
    "days": 7
}
```

#### Reply 
```json
[
    {
        "processid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "executorname": "my_executor_name",
        "message": "Error occurred during processing",
        "timestamp": 1640995200
    }
]
```

## User Management API

### Add User
* PayloadType: [**addusermsg**](../pkg/rpc/add_user_msg.go)
* Credentials: A valid Colony Private Key
* Returns: A [User object](../pkg/core/user.go)

#### Payload 
```json
{
    "msgtype": "addusermsg",
    "user": {
        "userid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "name": "john_doe",
        "email": "john@example.com",
        "phone": "+1234567890"
    }
}
```

#### Reply 
```json
{
    "userid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "name": "john_doe",
    "email": "john@example.com",
    "phone": "+1234567890"
}
```

### Get User
* PayloadType: [**getusermsg**](../pkg/rpc/get_user_msg.go)
* Credentials: A valid Colony Private Key
* Returns: A [User object](../pkg/core/user.go)

#### Payload 
```json
{
    "msgtype": "getusermsg",
    "colonyname": "my_colony_name",
    "name": "john_doe"
}
```

#### Reply 
```json
{
    "userid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "name": "john_doe",
    "email": "john@example.com",
    "phone": "+1234567890"
}
```

### List Users
* PayloadType: [**getusersmsg**](../pkg/rpc/get_users_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An array of [User objects](../pkg/core/user.go)

#### Payload 
```json
{
    "msgtype": "getusersmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
[
    {
        "userid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "name": "john_doe",
        "email": "john@example.com",
        "phone": "+1234567890"
    }
]
```

### Remove User
* PayloadType: [**removeusermsg**](../pkg/rpc/remove_user_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeusermsg",
    "colonyname": "my_colony_name",
    "name": "john_doe"
}
```

#### Reply 
```json
{}
```

## Snapshot API

### Create Snapshot
* PayloadType: [**createsnapshotmsg**](../pkg/rpc/create_snapshot_msg.go)
* Credentials: A valid Colony Private Key
* Returns: A [Snapshot object](../pkg/core/snapshot.go)

#### Payload 
```json
{
    "msgtype": "createsnapshotmsg",
    "colonyname": "my_colony_name",
    "label": "daily_snapshot",
    "name": "snapshot_2022_01_02"
}
```

#### Reply 
```json
{
    "snapshotid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "label": "daily_snapshot",
    "name": "snapshot_2022_01_02",
    "timestamp": "2022-01-02T12:00:00Z"
}
```

### Get Snapshot
* PayloadType: [**getsnapshotmsg**](../pkg/rpc/get_snapshot_msg.go)
* Credentials: A valid Colony Private Key
* Returns: A [Snapshot object](../pkg/core/snapshot.go)

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "getsnapshotmsg",
    "colonyname": "my_colony_name",
    // Use one of the following methods to identify the snapshot.
    // Method 1: Get by unique snapshot ID (highest priority).
    "snapshotid": "a-valid-sha256-hash-id",

    // Method 2: Get by unique name (used if snapshotid is empty).
    "name": "snapshot_2022_01_02"
}
```

#### Reply 
```json
{
    "snapshotid": "a-valid-sha256-hash-id",
    "colonyname": "my_colony_name",
    "label": "daily_snapshot",
    "name": "snapshot_2022_01_02",
    "timestamp": "2022-01-02T12:00:00Z"
}
```

### List Snapshots
* PayloadType: [**getsnapshotsmsg**](../pkg/rpc/get_snapshots_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An array of [Snapshot objects](../pkg/core/snapshot.go)

#### Payload 
```json
{
    "msgtype": "getsnapshotsmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
[
    {
        "snapshotid": "a-valid-sha256-hash-id",
        "colonyname": "my_colony_name",
        "label": "daily_snapshot",
        "name": "snapshot_2022_01_02",
        "timestamp": "2022-01-02T12:00:00Z"
    }
]
```

### Remove Snapshot
* PayloadType: [**removesnapshotmsg**](../pkg/rpc/remove_snapshot_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "removesnapshotmsg",
    "colonyname": "my_colony_name",
    // Use one of the following methods to identify the snapshot for deletion.
    // Method 1: Delete by unique snapshot ID (highest priority).
    "snapshotid": "a-valid-sha256-hash-id",

    // Method 2: Delete by unique name (used if snapshotid is empty).
    "name": "snapshot_2022_01_02"
}
```

#### Reply 
```json
{}
```

### Remove All Snapshots
* PayloadType: [**removeallsnapshotsmsg**](../pkg/rpc/remove_all_snapshots_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removeallsnapshotsmsg",
    "colonyname": "my_colony_name"
}
```

#### Reply 
```json
{}
```

## Server & Miscellaneous API

### Get Cluster Info
* PayloadType: [**getclustermsg**](../pkg/rpc/get_cluster_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: A [cluster Config object](../pkg/cluster/config.go)

#### Payload 
```json
{
    "msgtype": "getclustermsg"
}
```

#### Reply 
```json
{
    "nodes": [
        {
            "name": "server1",
            "host": "localhost",
            "apiport": 50080,
            "etcdclientport": 23100,
            "etcdpeerport": 24100,
            "relayport": 25100,
            "leader": true
        }
    ]
}
```

### Get Server Version
* PayloadType: [**versionmsg**](../pkg/rpc/version_msg.go)
* Credentials: None required
* Returns: A [version message](../pkg/rpc/version_msg.go)

#### Payload 
```json
{
    "msgtype": "versionmsg"
}
```

#### Reply 
```json
{
    "buildversion": "v1.0.0",
    "buildtime": "2022-01-02T10:00:00Z"
}
```

### Get System-Wide Statistics
* PayloadType: [**getstatisticsmsg**](../pkg/rpc/get_statistics_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: A [Statistics object](../pkg/core/statistics.go)

#### Payload 
```json
{
    "msgtype": "getstatisticsmsg"
}
```

#### Reply 
```json
{
    "colonies": 5,
    "executors": 25,
    "waitingprocesses": 10,
    "runningprocesses": 8,
    "successfulprocesses": 1500,
    "failedprocesses": 23,
    "waitingworkflows": 5,
    "runningworkflows": 2,
    "successfulworkflows": 150,
    "failedworkflows": 3
}
```

### Add Function
* PayloadType: [**addfunctionmsg**](../pkg/rpc/add_function_msg.go)
* Credentials: A valid Executor Private Key
* Returns: A [Function object](../pkg/core/function.go)

*Note: The functionid and executortype fields are set by the server and will be ignored if provided.*

#### Payload 
```json
{
    "msgtype": "addfunctionmsg",
    "fun": {
        "executorname": "my_executor_name",
        "colonyname": "my_colony_name",
        "funcname": "calculate_sum"
    }
}
```

#### Reply 
```json
{
    "functionid": "a-valid-sha256-hash-id",
    "executorname": "my_executor_name",
    "executortype": "test_executor_type",
    "colonyname": "my_colony_name",
    "funcname": "calculate_sum",
    "counter": 5,
    "minwaittime": 2.0,
    "maxwaittime": 3.0,
    "minexectime": 9.5,
    "maxexectime": 10.8,
    "avgwaittime": 2.5,
    "avgexectime": 10.1
}
```

### Get Functions
* PayloadType: [**getfunctionsmsg**](../pkg/rpc/get_functions_msg.go)
* Credentials: A valid Executor or Colony Private Key
* Returns: An array of [Function objects](../pkg/core/function.go)
* Comments: If `executorname` is not provided, the functions of all executors in the specified colony will be returned.

#### Payload 
*Note: Comments (//) are for documentation and must be removed before sending the request.*
```json
{
    "msgtype": "getfunctionsmsg",
    "colonyname": "my_colony_name",
    // This field is optional.
    // If provided, returns functions for a specific executor.
    // If omitted, returns all functions in the colony.
    "executorname": "my_executor"
}
```


#### Reply 
```json
[
    {
        "functionid": "a-valid-sha256-hash-id",
        "executorname": "my_executor_name",
        "executortype": "test_executor_type",
        "colonyname": "my_colony_name",
        "funcname": "calculate_sum",
        "counter": 5,
        "minwaittime": 2.0,
        "maxwaittime": 3.0,
        "minexectime": 9.5,
        "maxexectime": 10.8,
        "avgwaittime": 2.5,
        "avgexectime": 10.1
    }
]
```

### Remove Function
* PayloadType: [**removefunctionmsg**](../pkg/rpc/remove_function_msg.go)
* Credentials: A valid Executor Private Key
* Returns: An empty JSON object {}

#### Payload 
```json
{
    "msgtype": "removefunctionmsg",
    "functionid": "a-valid-sha256-hash-id"
}
```

#### Reply 
```json
{}
```

### ID Management

#### Change Colony ID
* PayloadType: [**changecolonyidmsg**](../pkg/rpc/change_colonyid_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: An empty JSON object {}

##### Payload 
```json
{
    "msgtype": "changecolonyidmsg",
    "colonyname": "my_colony_name",
    "colonyid": "a-new-colony-id-hash"
}
```

##### Reply 
```json
{}
```

#### Change Executor ID
* PayloadType: [**changeexecutoridmsg**](../pkg/rpc/change_executorid_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

##### Payload 
```json
{
    "msgtype": "changeexecutoridmsg",
    "colonyname": "my_colony_name",
    "executorid": "a-new-executor-id-hash"
}
```

##### Reply 
```json
{}
```

#### Change User ID
* PayloadType: [**changeuseridmsg**](../pkg/rpc/change_userid_msg.go)
* Credentials: A valid Colony Private Key
* Returns: An empty JSON object {}

##### Payload 
```json
{
    "msgtype": "changeuseridmsg",
    "colonyname": "my_colony_name",
    "userid": "a-new-user-id-hash"
}
```

##### Reply 
```json
{}
```

#### Change Server ID
* PayloadType: [**changeserveridmsg**](../pkg/rpc/change_serverid_msg.go)
* Credentials: A valid Server Owner Private Key
* Returns: An empty JSON object {}

##### Payload 
```json
{
    "msgtype": "changeserveridmsg",
    "serverid": "a-new-server-id-hash"
}
```

##### Reply 
```json
{}
```
