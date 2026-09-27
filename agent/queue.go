package agent

import (
	"fmt"
	"sync"
	"time"
)

// QueueStatus defines lifecycle state of queued prompts.
type QueueStatus string

const (
	QueueStatusQueued    QueueStatus = "queued"
	QueueStatusRunning   QueueStatus = "running"
	QueueStatusCompleted QueueStatus = "completed"
	QueueStatusFailed    QueueStatus = "failed"
	QueueStatusCancelled QueueStatus = "cancelled"
	QueueStatusPaused    QueueStatus = "paused"
)

// QueuedPrompt represents an individual task in the execution queue.
type QueuedPrompt struct {
	ID           string      `json:"id"`
	RunID        string      `json:"runId,omitempty"`
	Prompt       string      `json:"prompt"`
	Mode         string      `json:"mode"` // "plan" or "code"
	CreatedAt    time.Time   `json:"createdAt"`
	Status       QueueStatus `json:"status"`
	Dependencies []string    `json:"dependencies,omitempty"`
}

// QueueManager provides thread-safe prompt queue management and execution ordering.
type QueueManager struct {
	mu     sync.RWMutex
	items  []*QueuedPrompt
	paused bool
	broker *EventBroker
}

// NewQueueManager initializes QueueManager.
func NewQueueManager(broker *EventBroker) *QueueManager {
	return &QueueManager{
		items:  make([]*QueuedPrompt, 0),
		broker: broker,
	}
}

// Add appends a new prompt to the queue.
func (qm *QueueManager) Add(prompt, mode string, dependencies []string) *QueuedPrompt {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	id := fmt.Sprintf("q_%d_%d", time.Now().UnixNano(), len(qm.items)+1)
	if mode == "" {
		mode = "code"
	}

	item := &QueuedPrompt{
		ID:           id,
		Prompt:       prompt,
		Mode:         mode,
		CreatedAt:    time.Now(),
		Status:       QueueStatusQueued,
		Dependencies: dependencies,
	}

	qm.items = append(qm.items, item)
	qm.emitQueueEvent(EventType("queue.created"), item)
	return item
}

// Get returns a queued item by ID.
func (qm *QueueManager) Get(id string) (*QueuedPrompt, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	for _, item := range qm.items {
		if item.ID == id {
			return item, true
		}
	}
	return nil, false
}

// List returns all queued items in current order.
func (qm *QueueManager) List() []*QueuedPrompt {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	list := make([]*QueuedPrompt, len(qm.items))
	copy(list, qm.items)
	return list
}

// Reorder reorders the queue items by ID array.
func (qm *QueueManager) Reorder(ids []string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	idMap := make(map[string]*QueuedPrompt)
	for _, item := range qm.items {
		idMap[item.ID] = item
	}

	reordered := make([]*QueuedPrompt, 0, len(ids))
	for _, id := range ids {
		if item, ok := idMap[id]; ok {
			reordered = append(reordered, item)
			delete(idMap, id)
		}
	}

	for _, item := range idMap {
		reordered = append(reordered, item)
	}

	qm.items = reordered
	qm.emitQueueEvent(EventType("queue.updated"), qm.items)
}

// SetStatus updates the status of a queued prompt.
func (qm *QueueManager) SetStatus(id string, status QueueStatus) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, item := range qm.items {
		if item.ID == id {
			item.Status = status
			qm.emitQueueEvent(EventType("queue.updated"), item)
			break
		}
	}
}

// Remove deletes an item from the queue.
func (qm *QueueManager) Remove(id string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	filtered := make([]*QueuedPrompt, 0, len(qm.items))
	for _, item := range qm.items {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}
	qm.items = filtered
	qm.emitQueueEvent(EventType("queue.updated"), qm.items)
}

func (qm *QueueManager) emitQueueEvent(evtType EventType, data interface{}) {
	if qm.broker != nil {
		evt := Event{
			ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
			Type:      evtType,
			Timestamp: time.Now(),
			Data:      data,
		}
		qm.broker.Publish(evt)
	}
}
