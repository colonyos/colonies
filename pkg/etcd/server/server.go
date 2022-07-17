package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/davecgh/go-spew/spew"
	"go.etcd.io/etcd/server/v3/embed"
)

func main() {
	var nameFlag = flag.String("name", "", "give a name")
	var serverPortFlag = flag.String("serverport", "", "give a name")
	var clientPortFlag = flag.String("clientport", "", "give a name")
	flag.Parse()

	name := string(*nameFlag)
	serverPort := string(*serverPortFlag)
	clientPort := string(*clientPortFlag)

	fmt.Println(name)
	fmt.Println(serverPort)
	fmt.Println(clientPort)

	cfg := embed.NewConfig()
	cfg.LogLevel = "fatal"
	//cfg.LogLevel = "debug"
	cfg.Dir = name + "default.etcd"
	cfg.Name = name
	cfg.Logger = "zap"

	DefaultInitialAdvertisePeerURLs := "http://localhost:" + serverPort
	DefaultListenPeerURLs := "http://localhost:" + serverPort
	DefaultAdvertiseClientURLs := "http://localhost:" + clientPort
	DefaultListenClientURLs := "http://localhost:" + clientPort

	lpurl, _ := url.Parse(DefaultListenPeerURLs)           // --listen-peer-urls http://127.0.0.1:12380
	apurl, _ := url.Parse(DefaultInitialAdvertisePeerURLs) // --initial-advertise-peer-urls http://127.0.0.1:12380
	lcurl, _ := url.Parse(DefaultListenClientURLs)         // --listen-client-urls http://127.0.0.1:2379
	acurl, _ := url.Parse(DefaultAdvertiseClientURLs)      // --advertise-client-urls http://127.0.0.1:2379

	cfg.LPUrls = []url.URL{*lpurl}
	cfg.LCUrls = []url.URL{*lcurl}
	cfg.APUrls = []url.URL{*apurl}
	cfg.ACUrls = []url.URL{*acurl}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	cfg.InitialCluster = "s1=http://localhost:2380,s2=http://localhost:23800,s3=http://localhost:33800"
	cfg.InitialClusterToken = "etcd-cluster-1"

	fmt.Println(cfg.ClusterState)

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Server is ready!")
		for {
			//	spew.Dump(e.Server.Cluster())
			spew.Dump(e.Server.ID())
			spew.Dump(e.Server.Leader())
			time.Sleep(1 * time.Second)
		}
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Printf("Server took too long to start!")
	}
	log.Fatal(<-e.Err())
}
