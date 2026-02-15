package websocket

import (
	"sync"
)

type Client struct {
	ID   int64
	Send chan []byte
}

type Hub struct {
	clients map[int64]*Client
	sync.RWMutex
}

var GlobalHub = &Hub{
	clients: make(map[int64]*Client),
}

func (h *Hub) Register(uid int64, client *Client) {
	h.Lock()
	defer h.Unlock()
	h.clients[uid] = client
}

func (h *Hub) Unregister(uid int64) {
	h.Lock()
	defer h.Unlock()
	delete(h.clients, uid)
}

// GetAllClients 返回所有在线用户，用于广播
func (h *Hub) GetAllClients() []*Client {
	h.RLock()
	defer h.RUnlock()
	list := make([]*Client, 0, len(h.clients))
	for _, c := range h.clients {
		list = append(list, c)
	}
	return list
}
