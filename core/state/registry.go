// Registry projection of every recorded project. Per-project JSON files are
// projected as typed RegistryRow values. No positional column dependencies
// remain.
package state

// Registry returns the typed view of every recorded project, sorted by slug.
func (s *Store) Registry() ([]RegistryRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := s.listFiles()
	if err != nil {
		return nil, err
	}
	rows := make([]RegistryRow, 0, len(files))
	for _, f := range files {
		rec, err := s.loadFile(f)
		if err != nil {
			continue
		}
		rows = append(rows, recordToRow(rec))
	}
	return rows, nil
}

func recordToRow(rec Record) RegistryRow {
	p := rec.Project
	return RegistryRow{
		Slug:             p.Slug,
		AttachmentState:  rec.AttachmentState,
		Name:             p.Name,
		Dir:              p.Dir,
		Hostname:         p.Hostname,
		DocRoot:          p.DocRoot,
		ComposeProject:   p.ComposeProjectName,
		RuntimeNetwork:   p.RuntimeNetwork,
		DatabaseVolume:   p.DatabaseVolume,
		PHPVersion:       p.PHPVersion,
		MySQLDatabase:    p.MySQL.Database,
		MySQLPort:        p.MySQL.Port,
		PMAPort:          p.MySQL.PMAPort,
		WebNetworkAlias:  p.WebNetworkAlias,
		ContainerSummary: rec.Runtime.SummaryLine,
	}
}
