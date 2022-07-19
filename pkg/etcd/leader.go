package etcd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:23790", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer cli.Close()

	var name = flag.String("name", "", "give a name")
	flag.Parse()
	fmt.Println("My name is", string(*name))

	// // create a sessions to elect a Leader
	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	e := concurrency.NewElection(s, "/leader-election/")
	ctx := context.Background()
	// Elect a leader (or wait that the leader resign)
	if err := e.Campaign(ctx, "e"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Leader election for", *name)
	fmt.Println("Do some work in", *name)
	time.Sleep(5 * time.Second)
	if err := e.Resign(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Resign", *name)
}
