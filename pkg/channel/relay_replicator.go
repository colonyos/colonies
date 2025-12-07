package channel

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/cluster"
	log "github.com/sirupsen/logrus"
)

// ReplicationMessage types
const (
	ReplicateEntryType      = "replicate_entry"
	ReplicateChannelType    = "replicate_channel"
	ReplicateCleanupType    = "replicate_cleanup"
	ReplicateExecutorType   = "replicate_executor"
)

// ReplicationMessage is sent between servers via RelayServer
type ReplicationMessage struct {
	Type       string    `json:"type"`
	ChannelID  string    `json:"channelid,omitempty"`
	ProcessID  string    `json:"processid,omitempty"`
	ExecutorID string    `json:"executorid,omitempty"`
	Entry      *MsgEntry `json:"entry,omitempty"`
	Channel    *Channel  `json:"channel,omitempty"`
}

// RelayReplicator replicates channel operations using the cluster RelayServer
type RelayReplicator struct {
	relayServer *cluster.RelayServer
	localRouter *Router
}

// NewRelayReplicator creates a replicator that uses RelayServer for cluster communication
func NewRelayReplicator(relayServer *cluster.RelayServer, localRouter *Router) *RelayReplicator {
	r := &RelayReplicator{
		relayServer: relayServer,
		localRouter: localRouter,
	}

	// Start listening for incoming replication messages
	go r.handleIncoming()

	return r
}

// handleIncoming processes replication messages from other servers
func (r *RelayReplicator) handleIncoming() {
	incoming := r.relayServer.Receive()
	for msg := range incoming {
		var repMsg ReplicationMessage
		err := json.Unmarshal(msg.Data, &repMsg)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to unmarshal replication message")
			if msg.Done != nil {
				close(msg.Done)
			}
			continue
		}

		log.WithFields(log.Fields{"Type": repMsg.Type, "ChannelID": repMsg.ChannelID}).Debug("Received replication message")

		switch repMsg.Type {
		case ReplicateEntryType:
			// Ensure channel exists before replicating entry (race condition fix)
			if repMsg.Channel != nil {
				ch := &Channel{
					ID:          repMsg.Channel.ID,
					ProcessID:   repMsg.Channel.ProcessID,
					Name:        repMsg.Channel.Name,
					SubmitterID: repMsg.Channel.SubmitterID,
					ExecutorID:  repMsg.Channel.ExecutorID,
					Sequence:    0,
					Log:         make([]*MsgEntry, 0),
				}
				r.localRouter.CreateIfNotExists(ch)
			}
			if err := r.localRouter.ReplicateEntry(repMsg.ChannelID, repMsg.Entry); err != nil {
				log.WithFields(log.Fields{"Error": err, "ChannelID": repMsg.ChannelID}).Error("Failed to replicate entry")
			}
		case ReplicateChannelType:
			// Create a copy without the log entries
			// Use CreateIfNotExists to avoid re-replicating to peers
			ch := &Channel{
				ID:          repMsg.Channel.ID,
				ProcessID:   repMsg.Channel.ProcessID,
				Name:        repMsg.Channel.Name,
				SubmitterID: repMsg.Channel.SubmitterID,
				ExecutorID:  repMsg.Channel.ExecutorID,
				Sequence:    0,
				Log:         make([]*MsgEntry, 0),
			}
			if err := r.localRouter.CreateIfNotExists(ch); err != nil {
				log.WithFields(log.Fields{"Error": err, "ChannelID": ch.ID}).Error("Failed to replicate channel creation")
			}
		case ReplicateCleanupType:
			r.localRouter.CleanupProcess(repMsg.ProcessID)
		case ReplicateExecutorType:
			if err := r.localRouter.SetExecutorIDForProcess(repMsg.ProcessID, repMsg.ExecutorID); err != nil {
				log.WithFields(log.Fields{"Error": err, "ProcessID": repMsg.ProcessID}).Error("Failed to replicate executor assignment")
			}
		default:
			log.WithFields(log.Fields{"Type": repMsg.Type}).Warn("Unknown replication message type")
		}

		// Signal that processing is complete (if Done channel exists)
		if msg.Done != nil {
			close(msg.Done)
		}
	}
}

func (r *RelayReplicator) broadcast(msg *ReplicationMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"Type": msg.Type, "ChannelID": msg.ChannelID}).Debug("Broadcasting replication message")
	err = r.relayServer.Broadcast(data)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Type": msg.Type}).Error("Failed to broadcast replication message")
	}
	return err
}

func (r *RelayReplicator) ReplicateEntry(channel *Channel, entry *MsgEntry) error {
	// Include channel metadata so receiving server can create channel if it doesn't exist
	// This fixes the race condition where entry arrives before channel creation
	// Note: We create a shallow copy WITHOUT the Log to avoid data races
	channelMeta := &Channel{
		ID:          channel.ID,
		ProcessID:   channel.ProcessID,
		Name:        channel.Name,
		SubmitterID: channel.SubmitterID,
		ExecutorID:  channel.ExecutorID,
		Sequence:    0,
		Log:         nil, // Explicitly nil - we don't need the log for replication
	}
	msg := &ReplicationMessage{
		Type:      ReplicateEntryType,
		ChannelID: channel.ID,
		Entry:     entry,
		Channel:   channelMeta, // Include channel metadata for race condition handling
	}
	return r.broadcast(msg)
}

func (r *RelayReplicator) ReplicateChannel(channel *Channel) error {
	// Create a copy without the Log to avoid data races
	channelMeta := &Channel{
		ID:          channel.ID,
		ProcessID:   channel.ProcessID,
		Name:        channel.Name,
		SubmitterID: channel.SubmitterID,
		ExecutorID:  channel.ExecutorID,
		Sequence:    0,
		Log:         nil,
	}
	msg := &ReplicationMessage{
		Type:    ReplicateChannelType,
		Channel: channelMeta,
	}
	return r.broadcast(msg)
}

func (r *RelayReplicator) ReplicateCleanup(processID string) error {
	msg := &ReplicationMessage{
		Type:      ReplicateCleanupType,
		ProcessID: processID,
	}
	return r.broadcast(msg)
}

func (r *RelayReplicator) ReplicateExecutorAssignment(processID string, executorID string) error {
	msg := &ReplicationMessage{
		Type:       ReplicateExecutorType,
		ProcessID:  processID,
		ExecutorID: executorID,
	}
	return r.broadcast(msg)
}
