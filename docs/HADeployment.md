go run cmd/main.go server start --port 21001 --etcdname server2 --etcdhost localhost --etcdport 23101 --etcdpeerport 24101 --etcdcluster server1=localhost:24100,server2=localhost:24101

go run cmd/main.go server start --port 21000 --etcdname server1 --etcdhost localhost --etcdport 23100 --etcdpeerport 24100 --etcdcluster server1=localhost:24100,server2=localhost:24101
