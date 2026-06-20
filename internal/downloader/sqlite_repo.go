package downloader

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	metadataQuery := `
	CREATE TABLE IF NOT EXISTS download_metadata (
		id TEXT PRIMARY KEY,
		url TEXT NOT NULL,
		filename TEXT NOT NULL,
		total_size INTEGER,
		status INTEGER
	);
	`

	partQuery := `
	CREATE TABLE IF NOT EXISTS download_part_progress (
		id INTEGER PRIMARY KEY,
		download_id TEXT REFERENCES download_metadata(id),
		start_byte INTEGER NOT NULL,
		end_byte INTEGER NOT NULL,
		current_byte INTEGER NOT NULL,
		workers_id TEXT
	);
	`

	if _, err := conn.Exec(metadataQuery); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(partQuery); err != nil {
		return nil, err
	}
	return &SQLiteRepository{db: conn}, nil
}

func (r *SQLiteRepository) SaveDownload(state *DownloadState) error {
	query := `
	INSERT INTO download_metadata (id, url, filename, total_size, status)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET 
		url = excluded.url,
		filename = excluded.filename,
		total_size = excluded.total_size,
		status = excluded.status;
	`
	_, err := r.db.Exec(query, state.ID, state.URL, state.Filename, state.TotalSize, state.Status)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetDownload(id string) (*DownloadState, error) {
	state := &DownloadState{}
	var status int

	err := r.db.QueryRow("SELECT id, url, filename, total_size, status FROM download_metadata WHERE id = ?").Scan(&state.ID, &state.URL, &state.Filename, &state.TotalSize, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	state.Status = DownloadStatus(status)
	return state, nil
}

func (r *SQLiteRepository) GetIncompleteDownload() ([]*DownloadState, error) {
	rows, err := r.db.Query("SELECT id, url, filename, total_size, status FROM download_metadata WHERE status != ?", StateCompleted)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var state []*DownloadState
	for rows.Next() {
		currentState := &DownloadState{}

		if err := rows.Scan(
			&currentState.ID,
			&currentState.URL,
			&currentState.Filename,
			&currentState.TotalSize,
			&currentState.Status,
		); err != nil {
			return nil, err
		}

		state = append(state, currentState)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return state, nil
}

func (r *SQLiteRepository) GetAllDownloads() ([]*DownloadState, error) {
	rows, err := r.db.Query("SELECT id, url, filename, total_size, status FROM download_metadata")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var state []*DownloadState
	for rows.Next() {
		currentState := &DownloadState{}

		if err := rows.Scan(
			&currentState.ID,
			&currentState.URL,
			&currentState.Filename,
			&currentState.TotalSize,
			&currentState.Status,
		); err != nil {
			return nil, err
		}

		state = append(state, currentState)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return state, nil
}

func (r *SQLiteRepository) CreatePart(part *PartState) error {
	query := `
	INSERT INTO download_part_progress (id, download_id, start_byte, end_byte, current_byte, workers_id)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, part.ID, part.DownloadID, part.StartByte, part.EndByte, part.CurrentByte, part.WorkerID)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) UpdatePartsProgress(partID string, currentByte int64) error {
	query := `
	UPDATE download_part_progress SET current_byte = ? WHERE id = ?
	`
	_, err := r.db.Exec(query, currentByte, partID)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetParts(downloadID string) ([]*PartState, error) {
	rows, err := r.db.Query("SELECT id, download_id, start_byte, end_byte, current_byte, workers_id FROM download_part_progress WHERE download_id = ?")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var partState []*PartState
	for rows.Next() {
		currentPart := &PartState{}
		if err := rows.Scan(
			&currentPart.ID,
			&currentPart.DownloadID,
			&currentPart.StartByte,
			&currentPart.EndByte,
			&currentPart.CurrentByte,
			&currentPart.WorkerID,
		); err != nil {
			return nil, err
		}

		partState = append(partState, currentPart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return partState, nil
}
