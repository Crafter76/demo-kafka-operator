package mocks

import (
	"errors"
	"sync"
)

type MockKafkaClient struct {
	users map[string]bool
	mu    sync.RWMutex
	Calls int
}

func NewMockKafkaClient() *MockKafkaClient {
	return &MockKafkaClient{
		users: make(map[string]bool),
	}
}

func (m *MockKafkaClient) CreateUser(username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls++
	if m.users[username] {
		return errors.New("user already exists")
	}
	m.users[username] = true
	return nil
}

func (m *MockKafkaClient) HasUser(username string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[username]
}

func (m *MockKafkaClient) DeleteUser(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, username)
}
