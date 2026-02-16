package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestBus_Subscribe_Publish(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received Event
	var mu sync.Mutex

	bus.Subscribe(EventScanStarted, func(e Event) {
		mu.Lock()
		received = e
		mu.Unlock()
	})

	bus.Publish(NewEvent(EventScanStarted, ScanStartedData{
		ProjectPath: "/test",
		ProjectName: "test-project",
	}))

	// Give goroutine time to process
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if received.Type != EventScanStarted {
		t.Errorf("event type = %v, want %v", received.Type, EventScanStarted)
	}

	data, ok := received.Data.(ScanStartedData)
	if !ok {
		t.Fatal("expected ScanStartedData")
	}
	if data.ProjectName != "test-project" {
		t.Errorf("project name = %q, want %q", data.ProjectName, "test-project")
	}
}

func TestBus_SubscribeAll(t *testing.T) {
	bus := New()
	defer bus.Close()

	var count int
	var mu sync.Mutex

	bus.SubscribeAll(func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Publish(NewEvent(EventScanStarted, nil))
	bus.Publish(NewEvent(EventScanCompleted, nil))
	bus.Publish(NewEvent(EventLogMessage, nil))

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if count != 3 {
		t.Errorf("received %d events, want 3", count)
	}
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := New()
	defer bus.Close()

	var count1, count2 int
	var mu sync.Mutex

	bus.Subscribe(EventFindingDiscovered, func(e Event) {
		mu.Lock()
		count1++
		mu.Unlock()
	})
	bus.Subscribe(EventFindingDiscovered, func(e Event) {
		mu.Lock()
		count2++
		mu.Unlock()
	})

	bus.Publish(NewEvent(EventFindingDiscovered, FindingDiscoveredData{}))
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if count1 != 1 || count2 != 1 {
		t.Errorf("counts = %d, %d; want 1, 1", count1, count2)
	}
}

func TestNewEvent(t *testing.T) {
	e := NewEvent(EventScanFailed, ScanFailedData{Error: nil})
	if e.Type != EventScanFailed {
		t.Errorf("type = %v, want %v", e.Type, EventScanFailed)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}
