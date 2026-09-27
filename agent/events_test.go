package agent

import (
	"testing"
	"time"
)

func TestEventBroker_SubscribeAndPublish(t *testing.T) {
	broker := NewEventBroker()

	ch1, unsub1 := broker.Subscribe("run_1")
	defer unsub1()

	chAll, unsubAll := broker.Subscribe("")
	defer unsubAll()

	evt1 := Event{
		ID:        "evt_1",
		RunID:     "run_1",
		Type:      EventRunStarted,
		Timestamp: time.Now(),
		Data:      "test_data",
	}

	evt2 := Event{
		ID:        "evt_2",
		RunID:     "run_2",
		Type:      EventRunStarted,
		Timestamp: time.Now(),
		Data:      "test_data_2",
	}

	broker.Publish(evt1)
	broker.Publish(evt2)

	select {
	case received := <-ch1:
		if received.RunID != "run_1" {
			t.Errorf("Expected run_1, got %s", received.RunID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event on ch1")
	}

	select {
	case received := <-chAll:
		if received.RunID != "run_1" {
			t.Errorf("Expected run_1 first on chAll, got %s", received.RunID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first event on chAll")
	}

	select {
	case received := <-chAll:
		if received.RunID != "run_2" {
			t.Errorf("Expected run_2 second on chAll, got %s", received.RunID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second event on chAll")
	}
}
