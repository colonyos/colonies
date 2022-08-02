# Introduction
The Colonies server uses 3 internal ports. 
* The Etc client port (-etcdclientport) is used to by external clients to communicate with the Etcd server API. 
* The Etcd peer port (-etcdpeerport) is used for internal communication between Etcd servers.
* The Relayport port (-relayport) is used for internal communication between Colonies servers. 
* The API port port (-port) exposes the Colonies API. 

Note: For security reasons, only API port should be exposed externally on the Internet.

# Tutorial
Start 3 terminals and run the following command.Note that you first need to setup a PostgreSQL database and export the following environmental variables.

```console
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8
export LC_CTYPE=UTF-8
export TZ=Europe/Stockholm
export COLONIES_TLS="false"
export COLONIES_SERVERHOST="localhost"
export COLONIES_SERVERPORT="50080"
export COLONIES_MONITORPORT="21120"
export COLONIES_MONITORINTERVALL="1"
export COLONIES_SERVERID="039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103"
export COLONIES_SERVERPRVKEY="fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d"
export COLONIES_DBHOST="localhost"
export COLONIES_DBUSER="postgres"
export COLONIES_DBPORT="5432"
export COLONIES_DBPASSWORD="rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
export COLONIES_COLONYID="4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
export COLONIES_COLONYPRVKEY="ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514"
export COLONIES_RUNTIMEID="3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac"
export COLONIES_RUNTIMEPRVKEY="ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"
export COLONIES_RUNTIMETYPE="cli"
```

Type following command to start a TimescaleDB server.
```console
docker run -d --name timescaledb -p 5432:5432 -e POSTGRES_PASSWORD=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7 --restart unless-stopped timescale/timescaledb:latest-pg12
```

## Terminal 1
```console
colonies server start --port 50080 --relayport 25100 --etcdname server1 --etcdhost localhost --etcdclientport 23100 --etcdpeerport 24100 --initial-cluster server1=localhost:24100:25100:50080,server2=localhost:24101:25101:50081,server3=localhost:24102:25102:50082 --etcddatadir /tmp/colonies/test/etcd --insecure
```

## Terminal 2
```console
colonies server start --port 50081 --relayport 25101 --etcdname server2 --etcdhost localhost --etcdclientport 23101 --etcdpeerport 24101 --initial-cluster server1=localhost:24100:25100:50080,server2=localhost:24101:25101:50081,server3=localhost:24102:25102:50082 --etcddatadir /tmp/colonies/test/etcd --insecure
```

## Terminal 3 
```console
colonies server start --port 50082 --relayport 25102 --etcdname server3 --etcdhost localhost --etcdclientport 23102 --etcdpeerport 24102 --initial-cluster server1=localhost:24100:25100:50080,server2=localhost:24101:25101:50081,server3=localhost:24102:25102:50082 --etcddatadir /tmp/colonies/test/etcd --insecure
```

Test scripts for starting the servers above can also be found [here](./cluster-config).

## Check cluster status 
```console
colonies cluster info
```

Output:

```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
+---------+-----------+--------+
|  NAME   |   HOST    | LEADER |
+---------+-----------+--------+
| server1 | localhost | False  |
| server2 | localhost | False  |
| server3 | localhost | True   |
+---------+-----------+--------+
```

## Kill the leader 
Server 3 is the leader, kill it by pressing Ctrl-C. Notice the log message "INFO[0040] Colonies server came leader" in the Server 2 terminal window. 

```console
colonies cluster info
```

Output:

```console
INFO[0000] Starting a Colonies client                    Insecure=true ServerHost=localhost ServerPort=50080
+---------+-----------+---------+----------------+--------------+-----------+--------+
|  NAME   |   HOST    | APIPORT | ETCDCLIENTPORT | ETCDPEERPORT | RELAYPORT | LEADER |
+---------+-----------+---------+----------------+--------------+-----------+--------+
| server1 | localhost | 50080   | 23100          | 24100        | 25100     | True   |
| server2 | localhost | 50081   | 23100          | 24101        | 25101     | False  |
| server3 | localhost | 50082   | 23100          | 24102        | 25102     | False  |
+---------+-----------+---------+----------------+--------------+-----------+--------+
```

Server 2 is now the leader.
