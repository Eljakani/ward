package eventbus

import tea "github.com/charmbracelet/bubbletea"

// BusEventMsg wraps an EventBus event as a tea.Msg.
type BusEventMsg struct {
	Event Event
}

// Bridge forwards EventBus events into a Bubble Tea program.
type Bridge struct {
	bus     *EventBus
	program *tea.Program
}

// NewBridge creates a bridge between an EventBus and a tea.Program.
func NewBridge(bus *EventBus, program *tea.Program) *Bridge {
	return &Bridge{bus: bus, program: program}
}

// Start begins forwarding all events into the Bubble Tea program.
func (b *Bridge) Start() {
	b.bus.SubscribeAll(func(event Event) {
		b.program.Send(BusEventMsg{Event: event})
	})
}

// Stop closes the event bus.
func (b *Bridge) Stop() {
	b.bus.Close()
}
