package dht

import (
	"context"

	"github.com/colonyos/colonies/pkg/p2p"
)

// DHT defines the operations supported by a Distributed Hash Table in a Kademlia network.
type DHT interface {

	// Register adds a new node to the DHT network using a bootstrap node's address and a unique KademliaID.
	// The context allows for request cancellation and timeout control.
	RegisterNetwork(bootstrapNode p2p.Node, kademliaID string, ctx context.Context) error

	// RegisterNetworkWithAddr adds a new node to the DHT network using a bootstrap node's address and a unique KademliaID.
	RegisterNetworkWithAddr(bootstrapNodeAddr string, kademliaID string, ctx context.Context) error

	// FindContact retrieves information about a node identified by its KademliaID.
	// It returns a Contact structure with the node's information or an error if the node is not found.
	FindContact(kademliaID string, ctx context.Context) (Contact, error)

	// FindContacts finds up to 'count' number of contacts closest to the specified KademliaID.
	// It returns a slice of Contact structures or an error if the operation fails.
	FindContacts(kademliaID string, count int, ctx context.Context) ([]Contact, error)

	// Put stores a value in the DHT under a specified key. The key must adhere to a specific format
	// '/key1/key2/.../keyN' with 1 to 5 alphanumeric sublevels and no trailing slash. The 'replicationFactor'
	// determines the number of nodes across which the value is replicated.
	// The first key (key1) is referred to as the root key and is used by the DHT to find the correct node to store the value.
	Put(id string, prvKey string, key string, value string, replicationFactor int, ctx context.Context) error

	// Get retrieves all values stored under the subkeys of a given path from the DHT.
	// The key provided must follow the defined format '/key1/key2/.../keyN' with 1 to 5 alphanumeric sublevels
	// and no trailing slash. This method fetched data stored under the hierarchical key structure.
	// It returns a map of subkey-value pairs if successful or an error if the retrieval operation fails or if the
	// specified path does not exist.
	Get(id string, key string, replicationFactor int, ctx context.Context) ([]KV, error)

	// GetContact returns the contact information of the local DHT node.
	GetContact() Contact

	RegisterNode(id string, prvKey string, node *p2p.Node, ctx context.Context) error

	LookupNode(id string, name string, ctx context.Context) (*p2p.Node, error)

	// Shutdown gracefully stops the DHT node and releases all resources.
	Shutdown()
}
