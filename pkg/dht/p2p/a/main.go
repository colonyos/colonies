package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// chatServiceValidator validates that the value associated with a key is a valid peer ID.
type chatServiceValidator struct{}

// Validate conforms to the record.Validator interface.
// It checks if the given value is a valid peer ID for the chat service keys.
func (v chatServiceValidator) Validate(key string, value []byte) error {
	_, err := peer.Decode(string(value))
	if err != nil {
		return fmt.Errorf("invalid peer ID for key '%s'", key)
	}
	return nil
}

// Select conforms to the record.Validator interface.
// For this example, it just selects the first value, but you could implement more complex logic.
func (v chatServiceValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil // Simplest implementation: always select the first value.
}

func setupHostAndDHT(ctx context.Context) (host.Host, *dht.IpfsDHT, error) {
	// Create a new libp2p host
	h, err := libp2p.New()
	if err != nil {
		return nil, nil, err
	}

	nsValidator := record.NamespacedValidator{
		//"pk":   record.PublicKeyValidator{}, // Validator for public keys
		//"ipns": ipns.Validator{},
		"chat": chatServiceValidator{}, // Use your namespace here
	}

	// Create a DHT instance with the custom validator
	d, err := dht.New(ctx, h, dht.Validator(nsValidator))
	//	d, err := dht.New(ctx, h, dht.Validator(chatServiceValidator{}))
	if err != nil {
		h.Close()
		return nil, nil, err
	}

	// Bootstrap the DHT (in a real-world scenario, you would connect to known peers)
	if err := d.Bootstrap(ctx); err != nil {
		h.Close()
		return nil, nil, err
	}

	// Create a service key
	serviceKey := fmt.Sprintf("/myapp/chat/%s", "serviceName")

	// Advertise the service in the DHT
	if err := d.PutValue(ctx, serviceKey, []byte(h.ID())); err != nil {
		fmt.Println("err", err)
	}

	providers, err := d.GetValue(ctx, serviceKey)
	if err != nil {
		fmt.Println("err", err)
	}

	fmt.Println("providers", providers)

	return h, d, nil
}

func main() {
	ctx := context.Background()
	h, _, err := setupHostAndDHT(ctx)
	if err != nil {
		log.Fatalf("Failed to set up host and DHT: %s", err)
	}
	defer h.Close()

	// Your service advertisement and discovery logic here...

	<-ctx.Done()
}
