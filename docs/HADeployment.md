colonies server start --port 21000 --etcdname server1 --etcdhost localhost --etcdclientport 23100 --etcdpeerport 24100 --initial-cluster server1=localhost:24100,server2=localhost:24101

colonies server start --port 21001 --etcdname server2 --etcdhost localhost --etcdclientport 23101 --etcdpeerport 24101 --initial-cluster server1=localhost:24100,server2=localhost:24101

