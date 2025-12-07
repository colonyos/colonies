package channel

// Replicator defines the interface for replicating channel operations across servers
type Replicator interface {
	// ReplicateEntry sends a new entry to peer servers
	// Includes channel info so peers can create channel if it doesn't exist (race condition fix)
	ReplicateEntry(channel *Channel, entry *MsgEntry) error

	// ReplicateChannel sends channel creation to peer servers
	ReplicateChannel(channel *Channel) error

	// ReplicateCleanup notifies peers to cleanup a process's channels
	ReplicateCleanup(processID string) error

	// ReplicateExecutorAssignment notifies peers of executor assignment
	ReplicateExecutorAssignment(processID string, executorID string) error
}

// NoOpReplicator is a replicator that does nothing (for single-server setup)
type NoOpReplicator struct{}

func (r *NoOpReplicator) ReplicateEntry(channel *Channel, entry *MsgEntry) error {
	return nil
}

func (r *NoOpReplicator) ReplicateChannel(channel *Channel) error {
	return nil
}

func (r *NoOpReplicator) ReplicateCleanup(processID string) error {
	return nil
}

func (r *NoOpReplicator) ReplicateExecutorAssignment(processID string, executorID string) error {
	return nil
}

// InMemoryReplicator replicates to a list of peer routers (for testing)
type InMemoryReplicator struct {
	peers []*Router
}

// NewInMemoryReplicator creates a replicator for testing with in-memory peers
func NewInMemoryReplicator(peers []*Router) *InMemoryReplicator {
	return &InMemoryReplicator{
		peers: peers,
	}
}

func (r *InMemoryReplicator) ReplicateEntry(channel *Channel, entry *MsgEntry) error {
	for _, peer := range r.peers {
		// Ensure channel exists on peer before replicating entry
		peerChannel := &Channel{
			ID:          channel.ID,
			ProcessID:   channel.ProcessID,
			Name:        channel.Name,
			SubmitterID: channel.SubmitterID,
			ExecutorID:  channel.ExecutorID,
			Sequence:    0,
			Log:         make([]*MsgEntry, 0),
		}
		// Create channel if it doesn't exist (ignore ErrChannelExists)
		peer.CreateIfNotExists(peerChannel)

		if err := peer.ReplicateEntry(channel.ID, entry); err != nil {
			// In production, log error but continue to other peers
			continue
		}
	}
	return nil
}

func (r *InMemoryReplicator) ReplicateChannel(channel *Channel) error {
	for _, peer := range r.peers {
		// Create a copy of the channel for the peer
		peerChannel := &Channel{
			ID:          channel.ID,
			ProcessID:   channel.ProcessID,
			Name:        channel.Name,
			SubmitterID: channel.SubmitterID,
			ExecutorID:  channel.ExecutorID,
			Sequence:    0,
			Log:         make([]*MsgEntry, 0),
		}
		if err := peer.Create(peerChannel); err != nil {
			// May already exist from concurrent creation
			continue
		}
	}
	return nil
}

func (r *InMemoryReplicator) ReplicateCleanup(processID string) error {
	for _, peer := range r.peers {
		peer.CleanupProcess(processID)
	}
	return nil
}

func (r *InMemoryReplicator) ReplicateExecutorAssignment(processID string, executorID string) error {
	for _, peer := range r.peers {
		peer.SetExecutorIDForProcess(processID, executorID)
	}
	return nil
}
