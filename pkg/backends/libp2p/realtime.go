package libp2p

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"
)

// P2PRealtimeHandler handles realtime connections via libp2p pubsub
type P2PRealtimeHandler struct {
	pubsub *pubsub.PubSub
	ctx    context.Context
}

// NewP2PRealtimeHandler creates a new libp2p realtime handler
func NewP2PRealtimeHandler(pubsub *pubsub.PubSub) *P2PRealtimeHandler {
	return &P2PRealtimeHandler{
		pubsub: pubsub,
		ctx:    context.Background(),
	}
}

// HandleRealtimeRequest handles realtime subscription requests via pubsub
func (h *P2PRealtimeHandler) HandleRealtimeRequest(c backends.Context, jsonString string) {
	streamCtx, ok := c.(*StreamContext)
	if !ok {
		logrus.Error("Realtime handler requires libp2p stream context")
		c.String(400, "Invalid context for libp2p realtime handler")
		return
	}

	// Parse the subscription request
	var rpcMsg rpc.RPCMsg
	err := json.Unmarshal([]byte(jsonString), &rpcMsg)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse subscription request")
		c.String(400, "Invalid subscription request")
		return
	}

	// Determine the topic based on the RPC message type and content
	topic, err := h.getTopicForSubscription(&rpcMsg)
	if err != nil {
		logrus.WithError(err).Error("Failed to determine topic for subscription")
		c.String(400, err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"peer_id": streamCtx.GetPeerID(),
		"topic":   topic,
		"method":  rpcMsg.PayloadType,
	}).Info("Setting up libp2p pubsub subscription")

	// Join the topic
	topicHandle, err := h.pubsub.Join(topic)
	if err != nil {
		logrus.WithError(err).WithField("topic", topic).Error("Failed to join pubsub topic")
		c.String(400, fmt.Sprintf("Failed to join topic: %v", err))
		return
	}
	defer topicHandle.Close()

	// Subscribe to the topic
	subscription, err := topicHandle.Subscribe()
	if err != nil {
		logrus.WithError(err).WithField("topic", topic).Error("Failed to subscribe to pubsub topic")
		c.String(400, fmt.Sprintf("Failed to subscribe to topic: %v", err))
		return
	}
	defer subscription.Cancel()

	// Send confirmation that subscription is established
	confirmationMsg, err := rpc.CreateRPCReplyMsg("subscription_established", "{\"status\": \"subscribed\"}")
	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription confirmation")
		return
	}
	
	confirmationJSON, err := confirmationMsg.ToJSON()
	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription confirmation")
		return
	}

	streamCtx.String(200, "%s", confirmationJSON)

	// Start forwarding messages from pubsub to the stream
	for {
		msg, err := subscription.Next(h.ctx)
		if err != nil {
			logrus.WithError(err).WithField("topic", topic).Error("Pubsub subscription ended")
			return
		}

		// Skip our own messages
		if msg.ReceivedFrom == streamCtx.stream.Conn().LocalPeer() {
			continue
		}

		// Forward the message to the stream
		_, err = streamCtx.stream.Write(msg.Data)
		if err != nil {
			logrus.WithError(err).WithField("peer_id", streamCtx.GetPeerID()).Error("Failed to forward pubsub message to stream")
			return
		}
	}
}

// getTopicForSubscription determines the appropriate pubsub topic based on the subscription request
func (h *P2PRealtimeHandler) getTopicForSubscription(rpcMsg *rpc.RPCMsg) (string, error) {
	switch rpcMsg.PayloadType {
	case rpc.SubscribeProcessesPayloadType:
		var msg rpc.SubscribeProcessesMsg
		err := json.Unmarshal([]byte(rpcMsg.DecodePayload()), &msg)
		if err != nil {
			return "", fmt.Errorf("failed to parse subscribe processes message: %w", err)
		}
		return fmt.Sprintf("/colonies/%s/processes", msg.ColonyName), nil
		
	case rpc.SubscribeProcessPayloadType:
		var msg rpc.SubscribeProcessMsg
		err := json.Unmarshal([]byte(rpcMsg.DecodePayload()), &msg)
		if err != nil {
			return "", fmt.Errorf("failed to parse subscribe process message: %w", err)
		}
		return fmt.Sprintf("/colonies/%s/process/%s", msg.ColonyName, msg.ProcessID), nil
		
	default:
		return "", fmt.Errorf("unsupported subscription type: %s", rpcMsg.PayloadType)
	}
}

// PublishProcessUpdate publishes a process update to the appropriate topic
func (h *P2PRealtimeHandler) PublishProcessUpdate(process *core.Process) error {
	// Create RPC reply message with the process
	processJSON, err := process.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal process: %w", err)
	}
	
	rpcReply, err := rpc.CreateRPCReplyMsg("process_update", processJSON)
	if err != nil {
		return fmt.Errorf("failed to create RPC reply: %w", err)
	}
	replyJSON, err := rpcReply.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal RPC reply: %w", err)
	}

	// Publish to colony-wide topic
	colonyTopic := fmt.Sprintf("/colonies/%s/processes", process.FunctionSpec.Conditions.ColonyName)
	if err := h.publishToTopic(colonyTopic, []byte(replyJSON)); err != nil {
		logrus.WithError(err).WithField("topic", colonyTopic).Error("Failed to publish to colony topic")
	}

	// Publish to process-specific topic
	processTopic := fmt.Sprintf("/colonies/%s/process/%s", process.FunctionSpec.Conditions.ColonyName, process.ID)
	if err := h.publishToTopic(processTopic, []byte(replyJSON)); err != nil {
		logrus.WithError(err).WithField("topic", processTopic).Error("Failed to publish to process topic")
	}

	return nil
}

// publishToTopic publishes a message to a specific pubsub topic
func (h *P2PRealtimeHandler) publishToTopic(topic string, data []byte) error {
	topicHandle, err := h.pubsub.Join(topic)
	if err != nil {
		return fmt.Errorf("failed to join topic %s: %w", topic, err)
	}
	defer topicHandle.Close()
	
	return topicHandle.Publish(h.ctx, data)
}