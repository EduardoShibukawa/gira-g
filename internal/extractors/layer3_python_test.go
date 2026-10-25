package extractors

import (
	"context"
	"testing"
)

func TestPythonASTExtractor(t *testing.T) {
	ext := NewPythonASTExtractor()

	if !ext.CanExtract("event_outbox/store.py") {
		t.Fatalf("expected CanExtract to match event_outbox/store.py")
	}
	if ext.CanExtract("tests/__pycache__/test.py") {
		t.Fatalf("expected CanExtract to ignore __pycache__")
	}

	pySource := `import threading
import time
from typing import Dict, List, Optional
from .models import Event, EventStatus

class OutboxStore:
    def __init__(self):
        self._lock = threading.Lock()
        self._events = {}

    def append(self, event: Event) -> Event:
        with self._lock:
            self._events[event.event_id] = event
            return event

    def acquire_next(self, worker_id: str, lease_seconds: float = 30.0) -> Optional[Event]:
        return None
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      "event_outbox/store.py",
		Content:   []byte(pySource),
	}

	nodes, _, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 class node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.Name != "OutboxStore" {
		t.Errorf("expected node name OutboxStore, got %s", node.Name)
	}

	methods, ok := node.Metadata["methods"].([]map[string]any)
	if !ok || len(methods) != 3 {
		t.Fatalf("expected 3 methods (__init__, append, acquire_next), got %v", node.Metadata["methods"])
	}

	appendMethod := methods[1]
	if appendMethod["name"] != "append" {
		t.Errorf("expected method append, got %v", appendMethod["name"])
	}
	if appendMethod["return_type"] != "Event" {
		t.Errorf("expected return_type Event, got %v", appendMethod["return_type"])
	}
}
