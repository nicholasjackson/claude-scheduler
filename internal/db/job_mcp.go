package db

// GetMCPServersForJob returns the MCP servers associated with a job.
func (s *Store) GetMCPServersForJob(jobID string) ([]MCPServer, error) {
	rows, err := s.db.Query(
		`SELECT m.id, m.name, m.type, m.url, m.command, m.args, m.env, m.headers
		 FROM mcp_servers m
		 INNER JOIN job_mcp_servers jm ON jm.mcp_server_id = m.id
		 WHERE jm.job_id = ?
		 ORDER BY m.name`,
		jobID,
	)
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

// SetJobMCPServers replaces all MCP server associations for a job.
func (s *Store) SetJobMCPServers(jobID string, serverIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM job_mcp_servers WHERE job_id = ?", jobID); err != nil {
		return err
	}

	for _, sid := range serverIDs {
		if _, err := tx.Exec(
			"INSERT INTO job_mcp_servers (job_id, mcp_server_id) VALUES (?, ?)",
			jobID, sid,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
