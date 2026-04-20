package gateway

import (
	"sync"
)

// MemoryApprovals is an in-process approval decision store for the gateway HTTP API.
// Production stacks should persist via ApprovalRequest status in etcd.
type MemoryApprovals struct {
	mu sync.Mutex
	m  map[string]string // id -> Allowed|Denied
}

// NewMemoryApprovals constructs the store.
func NewMemoryApprovals() *MemoryApprovals {
	return &MemoryApprovals{m: make(map[string]string)}
}

func (m *MemoryApprovals) Set(id, decision string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[id] = decision
}

func (m *MemoryApprovals) Get(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.m[id]
	return v, ok
}
