package eventbus

import "sync"

// Handler is a function that processes an event.
type Handler func(Event)

// EventBus provides a decoupled publish-subscribe mechanism.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]Handler
	allHandlers []Handler
	closed      bool
}

// New creates a new EventBus.
func New() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]Handler),
	}
}

// Subscribe registers a handler for a specific event type.
func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// SubscribeAll registers a handler that receives every event.
func (b *EventBus) SubscribeAll(handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.allHandlers = append(b.allHandlers, handler)
}

// Publish sends an event to all matching subscribers synchronously.
func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, h := range b.allHandlers {
		h(event)
	}
	for _, h := range b.subscribers[event.Type] {
		h(event)
	}
}

// Close marks the bus as closed; further Publish calls are no-ops.
func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
}
