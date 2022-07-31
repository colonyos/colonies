#!/usr/bin/env bash

colonies server start --port 50080 --relayport 25100 --etcdname server1 --etcdhost localhost --etcdclientport 23100 --etcdpeerport 24100 --initial-cluster server1=localhost:24100:25100:50080,server2=localhost:24101:25101:50081,server3=localhost:24102:25102:50082 --etcddatadir /tmp/colonies/test/etcd --insecure
