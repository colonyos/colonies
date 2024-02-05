package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/record"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new libp2p Host
	h, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %s", err)
	}
	defer h.Close()

	// Create a custom validator that includes namespaced validation logic
	customValidator := &record.ValidChecker{
		Validator: map[string]record.Validator{
			"utf8": &utf8Validator{},
		},
		Selector: map[string]record.Selector{
			"utf8": &record.CurrentSelector{},
		},
	}

	// Create a new DHT instance with the custom validator
	d, err := dht.New(ctx, h, dht.Validator(customValidator))
	if err != nil {
		log.Fatalf("Failed to create DHT with custom validator: %s", err)
	}

	// Bootstrap the DHT
	if err := d.Bootstrap(ctx); err != nil {
		log.Fatalf("Failed to bootstrap DHT: %s", err)
	}

	// Define the key and value to store in the DHT
	key := "/utf8/validKey"
	value := "Hello, world!"

	// Store the value in the DHT
	if err := storeValue(ctx, d, key, value); err != nil {
		log.Fatalf("Failed to store value: %s", err)
	}

	// Retrieve the value from the DHT
	retrievedValue, err := d.GetValue(ctx, key)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %s", err)
	}
	fmt.Printf("Retrieved value: %s\n", string(retrievedValue))
}

// utf8Validator is a custom validator that checks for valid UTF-8 encoding.
type utf8Validator struct{}

func (v *utf8Validator) Validate(key string, value []byte) error {
	if !utf8.Valid(value) {
		return fmt.Errorf("invalid UTF-8 value for key %s", key)
	}
	return nil
}

func (v *utf8Validator) Select(key string, values [][]byte) (int, error) {
	return 0, nil // Always select the first value
}

// storeValue stores a key-value pair in the DHT.
func storeValue(ctx context.Context, d *dht.IpfsDHT, key, value string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return d.PutValue(ctx, key, []byte(value))
}
