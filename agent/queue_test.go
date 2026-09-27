package agent

import (
	"testing"
)

func TestQueueManager_AddListReorderRemove(t *testing.T) {
	qm := NewQueueManager(nil)

	item1 := qm.Add("Task 1", "code", nil)
	item2 := qm.Add("Task 2", "plan", nil)
	item3 := qm.Add("Task 3", "code", nil)

	items := qm.List()
	if len(items) != 3 {
		t.Fatalf("Expected 3 queued items, got %d", len(items))
	}

	// Reorder items: item3, item1, item2
	qm.Reorder([]string{item3.ID, item1.ID, item2.ID})

	reordered := qm.List()
	if len(reordered) != 3 {
		t.Fatalf("Expected 3 reordered items, got %d", len(reordered))
	}

	if reordered[0].ID != item3.ID || reordered[1].ID != item1.ID || reordered[2].ID != item2.ID {
		t.Errorf("Reorder failed: %+v", reordered)
	}

	// Status machine update
	qm.SetStatus(item3.ID, QueueStatusRunning)
	if qm.List()[0].Status != QueueStatusRunning {
		t.Errorf("Expected status running for item3")
	}

	// Remove item
	qm.Remove(item1.ID)
	if len(qm.List()) != 2 {
		t.Errorf("Expected 2 items after removal, got %d", len(qm.List()))
	}
}
