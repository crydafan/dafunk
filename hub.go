package main

import (
	"log"
	"sync"
)

type Hub struct {
	clients map[chan []byte]struct{}
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan []byte]struct{}),
	}
}

func (h *Hub) AddClient(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ch] = struct{}{}
	log.Printf("Client joined, total: %d", len(h.clients))
}

func (h *Hub) RemoveClient(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ch)
	close(ch)
	log.Printf("Client left, total: %d", len(h.clients))
}

func (h *Hub) Broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		// Non-blocking send
		select {
		case ch <- data:
		default:
		}

	}
}
