package db

import (
	"fmt"

	"github.com/google/uuid"
)

// MCPServer represents an MCP server configuration persisted in the database.
type MCPServer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`    // "http" or "stdio"
	URL     string `json:"url"`     // for http type
	Command string `json:"command"` // for stdio type
	Args    string `json:"args"`    // JSON array string
	Env     string `json:"env"`     // JSON object string
	Headers string `json:"headers"` // JSON object string
}

var validMCPTypes = map[string]bool{
	"http":  true,
	"stdio": true,
}

func validateMCPServer(s MCPServer) error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !validMCPTypes[s.Type] {
		return fmt.Errorf("invalid type: %s (must be http or stdio)", s.Type)
	}
	if s.Type == "http" && s.URL == "" {
		return fmt.Errorf("url is required for http type")
	}
	if s.Type == "stdio" && s.Command == "" {
		return fmt.Errorf("command is required for stdio type")
	}
	return nil
}

// GetMCPServers returns all MCP servers ordered by name.
func (s *Store) GetMCPServers() ([]MCPServer, error) {
	rows, err := s.db.Query(
		"SELECT id, name, type, url, command, args, env, headers FROM mcp_servers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []MCPServer{}
	for rows.Next() {
		var srv MCPServer
		if err := rows.Scan(&srv.ID, &srv.Name, &srv.Type, &srv.URL, &srv.Command,
			&srv.Args, &srv.Env, &srv.Headers); err != nil {
			return nil, err
		}
		servers = append(servers, srv)
	}
	return servers, rows.Err()
}

// GetMCPServer returns a single MCP server by ID.
func (s *Store) GetMCPServer(id string) (MCPServer, error) {
	var srv MCPServer
	err := s.db.QueryRow(
		"SELECT id, name, type, url, command, args, env, headers FROM mcp_servers WHERE id = ?",
		id,
	).Scan(&srv.ID, &srv.Name, &srv.Type, &srv.URL, &srv.Command,
		&srv.Args, &srv.Env, &srv.Headers)
	return srv, err
}

// CreateMCPServer inserts a new MCP server. It assigns a UUID if ID is empty.
func (s *Store) CreateMCPServer(srv MCPServer) (MCPServer, error) {
	if err := validateMCPServer(srv); err != nil {
		return srv, err
	}
	if srv.ID == "" {
		srv.ID = uuid.New().String()
	}
	if srv.Args == "" {
		srv.Args = "[]"
	}
	if srv.Env == "" {
		srv.Env = "{}"
	}
	if srv.Headers == "" {
		srv.Headers = "{}"
	}

	_, err := s.db.Exec(
		`INSERT INTO mcp_servers (id, name, type, url, command, args, env, headers)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		srv.ID, srv.Name, srv.Type, srv.URL, srv.Command, srv.Args, srv.Env, srv.Headers,
	)
	return srv, err
}

// UpdateMCPServer updates an existing MCP server.
func (s *Store) UpdateMCPServer(srv MCPServer) (MCPServer, error) {
	if err := validateMCPServer(srv); err != nil {
		return srv, err
	}

	result, err := s.db.Exec(
		`UPDATE mcp_servers SET name=?, type=?, url=?, command=?, args=?, env=?, headers=?
		 WHERE id=?`,
		srv.Name, srv.Type, srv.URL, srv.Command, srv.Args, srv.Env, srv.Headers, srv.ID,
	)
	if err != nil {
		return srv, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return srv, fmt.Errorf("mcp server not found: %s", srv.ID)
	}
	return srv, nil
}

// DeleteMCPServer removes an MCP server by ID.
func (s *Store) DeleteMCPServer(id string) error {
	result, err := s.db.Exec("DELETE FROM mcp_servers WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("mcp server not found: %s", id)
	}
	return nil
}
