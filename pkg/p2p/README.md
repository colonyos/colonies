# Introduction
This repo contains the Colonies P2P network stack based in [libp2p](https://libp2p.io/).

Colonies uses it own DHT (Kademlia) implementation. Here is example code how to setup a DHT network 
constisting of two nodes.

```go
dht1, err := dht.CreateDHT(4001, "dht1")
dht2, err := dht.CreateDHT(4002, "dht2")
```

Register dht2 with dht1. 

```go
err = dht2.RegisterNetworkWithAddr(c.Node.Addr, c.ID.String(), ctx)
```
When using libp2p, the adresses will look like this:
**/ip4/10.0.0.201/tcp/4001/p2p/12D3KooWMkeo4JoJkf2e5CQgkZy55EkD4gXR8M2oW9WtvMDBS17K**

To publish key-value pair in the network, we must have a valid ECDSA private key.
```go
crypto := crypto.CreateCrypto()
prvKey, err := crypto.GeneratePrivateKey()
id, err := crypto.GenerateID(prvKey)
```
