package main

import "context"

func getClientFromContext(ctx context.Context) *Client {
	if client, ok := ctx.Value("client").(*Client); ok {
		return client
	}
	return nil
}

func (cfg *serverConfig) updateClientID(oldID, newID string) {
	cfg.ClientsMu.Lock()
	defer cfg.ClientsMu.Unlock()

	if client, exists := cfg.Clients[oldID]; exists {
		delete(cfg.Clients, oldID)
		client.ID = newID
		client.IsAuthed = true
		cfg.Clients[newID] = client
	}
}
