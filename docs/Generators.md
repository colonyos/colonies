# Generators
Generators automatically spawn workflows and can be used to create recurring workflows, for example workflows that automatically process data in a database once new data becomes available. In the current implementation, new workflows are automatically submitted by the Colonies server based on two triggers:

1. The first trigger is a counter mechanism. A workflow is automatically submitted if the counter exceeds a trigger threshold (counter>=threshold). A third-party server may for example increase the counter (via the Colonies API) to indicate that new data has been added. 
A new workflow will then automatically be triggered to process the newly added data when the threshold is exceeded.
2. The second trigger is a timeout mechanism. A workflow is automatically triggered if a workflow has not been generated for a certain amount of time, even if counter<threshold. However, the counter needs to be greater than 1 (counter>1) for a workflow to be triggered. 
The rationale is that a workflow should only be triggered if new data has been added.

Note if many Colonies servers run in a cluster and are connected to the same PostrgreSQL database, one of the Colonies server is then elected as a leader. Only the leader can execute generators. A new leader is automatically elected if the current leader becomes unavailable.

## Add a generator
```console
colonies generator add --spec examples/workflow.json --name testgenerator --timeout 5 --trigger 5
```
Output:
```
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
INFO[0000] Generator added                               GeneratorID=97cf378e0145fc5ff5e1c7bb8aa088f890e12cf66c87c543b2b90e2f4ee21390
```

This will add a new generator that will automatically submit a workflow after 5 seconds assuming the counter has been increased by one, or of the counter exceeds 5.

## Increase generator counter
```console
colonies generator inc --generatorid 97cf378e0145fc5ff5e1c7bb8aa088f890e12cf66c87c543b2b90e2f4ee21390
```

The command above will increase the counter for the specified generator.
