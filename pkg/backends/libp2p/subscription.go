package libp2p

import (
	"fmt"
	"sync"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/sirupsen/logrus"
)

// SubscriptionController manages libp2p-based subscriptions
type SubscriptionController struct {
	eventHandler backends.RealtimeEventHandler
	subscriptions map[string][]*backends.RealtimeSubscription
	mu sync.RWMutex
}

// NewSubscriptionController creates a new libp2p subscription controller
func NewSubscriptionController(eventHandler backends.RealtimeEventHandler) backends.RealtimeSubscriptionController {
	return &SubscriptionController{
		eventHandler: eventHandler,
		subscriptions: make(map[string][]*backends.RealtimeSubscription),
	}
}

// AddProcessesSubscriber adds a subscription for all processes of a certain type
func (s *SubscriptionController) AddProcessesSubscriber(executorID string, subscription *backends.RealtimeSubscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	logrus.WithFields(logrus.Fields{
		"executor_id": executorID,
		"executor_type": subscription.ExecutorType,
		"state": subscription.State,
	}).Debug("Adding libp2p processes subscriber")
	
	key := s.getSubscriptionKey(executorID, subscription.ExecutorType, subscription.State, "")
	s.subscriptions[key] = append(s.subscriptions[key], subscription)
	
	return nil
}

// AddProcessSubscriber adds a subscription for a specific process
func (s *SubscriptionController) AddProcessSubscriber(executorID string, process *core.Process, subscription *backends.RealtimeSubscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	logrus.WithFields(logrus.Fields{
		"executor_id": executorID,
		"process_id": process.ID,
		"executor_type": subscription.ExecutorType,
		"state": subscription.State,
	}).Debug("Adding libp2p process subscriber")
	
	key := s.getSubscriptionKey(executorID, subscription.ExecutorType, subscription.State, process.ID)
	s.subscriptions[key] = append(s.subscriptions[key], subscription)
	
	return nil
}

// NotifySubscribers notifies all relevant subscribers about a process event
func (s *SubscriptionController) NotifySubscribers(process *core.Process) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Notify general subscribers (all processes of this type)
	generalKey := s.getSubscriptionKey("", process.FunctionSpec.Conditions.ExecutorType, process.State, "")
	if subscriptions, exists := s.subscriptions[generalKey]; exists {
		s.notifySubscriptionList(subscriptions, process)
	}
	
	// Notify specific process subscribers
	processKey := s.getSubscriptionKey("", process.FunctionSpec.Conditions.ExecutorType, process.State, process.ID)
	if subscriptions, exists := s.subscriptions[processKey]; exists {
		s.notifySubscriptionList(subscriptions, process)
	}
	
	// Notify executor-specific subscribers
	for key, subscriptions := range s.subscriptions {
		if s.keyMatchesProcess(key, process) {
			s.notifySubscriptionList(subscriptions, process)
		}
	}
}

// notifySubscriptionList sends process updates to a list of subscriptions
func (s *SubscriptionController) notifySubscriptionList(subscriptions []*backends.RealtimeSubscription, process *core.Process) {
	for _, subscription := range subscriptions {
		if subscription.Connection != nil && subscription.Connection.IsOpen() {
			// For libp2p, we can send the process data directly as JSON
			data, err := process.ToJSON()
			if err != nil {
				logrus.WithError(err).Error("Failed to marshal process to JSON")
				continue
			}
			
			err = subscription.Connection.WriteMessage(subscription.MsgType, []byte(data))
			if err != nil {
				logrus.WithError(err).Error("Failed to send process update via libp2p")
				// Close the connection if it failed
				subscription.Connection.Close()
			}
		}
	}
}

// getSubscriptionKey creates a unique key for subscription storage
func (s *SubscriptionController) getSubscriptionKey(executorID, executorType string, state int, processID string) string {
	if processID == "" {
		return fmt.Sprintf("%s_%s_%d", executorID, executorType, state)
	}
	return fmt.Sprintf("%s_%s_%d_%s", executorID, executorType, state, processID)
}

// keyMatchesProcess checks if a subscription key matches a process
func (s *SubscriptionController) keyMatchesProcess(key string, process *core.Process) bool {
	// This is a simplified implementation
	// In a real implementation, you'd parse the key and match against process fields
	return false // For now, rely on the explicit key matching above
}

// RemoveSubscription removes a subscription (cleanup method)
func (s *SubscriptionController) RemoveSubscription(executorID string, subscription *backends.RealtimeSubscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Find and remove the subscription from all relevant keys
	for key, subscriptions := range s.subscriptions {
		for i, sub := range subscriptions {
			if sub == subscription {
				// Remove this subscription
				s.subscriptions[key] = append(subscriptions[:i], subscriptions[i+1:]...)
				
				// Clean up empty subscription lists
				if len(s.subscriptions[key]) == 0 {
					delete(s.subscriptions, key)
				}
				
				logrus.WithField("key", key).Debug("Removed libp2p subscription")
				break
			}
		}
	}
	
	return nil
}

// GetSubscriptionCount returns the number of active subscriptions (for monitoring)
func (s *SubscriptionController) GetSubscriptionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, subscriptions := range s.subscriptions {
		count += len(subscriptions)
	}
	return count
}

// Compile-time check that SubscriptionController implements backends.RealtimeSubscriptionController
var _ backends.RealtimeSubscriptionController = (*SubscriptionController)(nil)