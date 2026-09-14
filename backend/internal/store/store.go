package store

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    _ "modernc.org/sqlite"

    "media-sequencer/backend/internal/model"
)

type Store struct {
	DB *sql.DB
	mu sync.RWMutex
}

func New(dbPath string) (*Store, error) {
    dir := filepath.Dir(dbPath)

    if dir != "." {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return nil, err
        }
    }

    db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	store := &Store{
		DB: db,
	}

	if err := store.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	if err := store.seedData(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) createTables() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS windows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			window_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			url TEXT NOT NULL,
			duration INTEGER NOT NULL,
			position INTEGER NOT NULL,
			FOREIGN KEY(window_id) REFERENCES windows(id)
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS sync_state (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			media_id INTEGER NOT NULL,
			started_at INTEGER NOT NULL,
			duration INTEGER NOT NULL
		)
		`,
	}

	for _, query := range queries {
		if _, err := s.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) seedData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int

	err := s.DB.QueryRow(
		"SELECT COUNT(*) FROM windows",
	).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	windows := []string{
		"Window 1",
		"Window 2",
		"Window 3",
	}

	for _, name := range windows {
		result, err := s.DB.Exec(
			"INSERT INTO windows (name) VALUES (?)",
			name,
		)

		if err != nil {
			return err
		}

		windowID, err := result.LastInsertId()

		if err != nil {
			return err
		}

		media := []struct {
			name      string
			mediaType string
			url       string
			duration  int
			position  int
		}{
			{
				name:      "Sample Image",
				mediaType: "image",
				url:       "https://picsum.photos/800/450",
				duration:  10,
				position:  1,
			},
			{
				name:      "Sample Video",
				mediaType: "video",
				url:       "https://www.w3schools.com/html/mov_bbb.mp4",
				duration:  20,
				position:  2,
			},
			{
				name:      "Blank",
				mediaType: "blank",
				url:       "",
				duration:  5,
				position:  3,
			},
		}

		for _, item := range media {
			_, err := s.DB.Exec(`
				INSERT INTO media
				(window_id, name, type, url, duration, position)
				VALUES (?, ?, ?, ?, ?, ?)
			`,
				windowID,
				item.name,
				item.mediaType,
				item.url,
				item.duration,
				item.position,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) GetWindows() ([]model.Window, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.DB.Query(`
		SELECT id, name
		FROM windows
		ORDER BY id
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var windows []model.Window

	for rows.Next() {
		var window model.Window

		if err := rows.Scan(
			&window.ID,
			&window.Name,
		); err != nil {
			return nil, err
		}

		playlist, err := s.getPlaylist(window.ID)

		if err != nil {
			return nil, err
		}

		window.Playlist = playlist

		windows = append(windows, window)
	}

	return windows, rows.Err()
}

func (s *Store) getPlaylist(windowID int) ([]model.Media, error) {
	rows, err := s.DB.Query(`
		SELECT id, window_id, name, type, url, duration, position
		FROM media
		WHERE window_id = ?
		ORDER BY position ASC, id ASC
	`, windowID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var playlist []model.Media

	for rows.Next() {
		var media model.Media

		if err := rows.Scan(
			&media.ID,
			&media.WindowID,
			&media.Name,
			&media.Type,
			&media.URL,
			&media.Duration,
			&media.Position,
		); err != nil {
			return nil, err
		}

		playlist = append(playlist, media)
	}

	return playlist, rows.Err()
}

func (s *Store) AddMedia(
	windowID int,
	name string,
	mediaType string,
	url string,
	duration int,
) (*model.Media, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int

	err := s.DB.QueryRow(
		"SELECT COUNT(*) FROM media WHERE window_id = ?",
		windowID,
	).Scan(&count)

	if err != nil {
		return nil, err
	}

	position := count + 1

	result, err := s.DB.Exec(`
		INSERT INTO media
		(window_id, name, type, url, duration, position)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		windowID,
		name,
		mediaType,
		url,
		duration,
		position,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &model.Media{
		ID:       int(id),
		WindowID: windowID,
		Name:     name,
		Type:     mediaType,
		URL:      url,
		Duration: duration,
		Position: position,
	}, nil
}

func (s *Store) GetMediaByID(id int) (*model.Media, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var media model.Media

	err := s.DB.QueryRow(`
		SELECT id, window_id, name, type, url, duration, position
		FROM media
		WHERE id = ?
	`, id).Scan(
		&media.ID,
		&media.WindowID,
		&media.Name,
		&media.Type,
		&media.URL,
		&media.Duration,
		&media.Position,
	)

	if err != nil {
		return nil, err
	}

	return &media, nil
}

func (s *Store) SetSync(mediaID int, duration int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.DB.Exec(`
		INSERT INTO sync_state (id, media_id, started_at, duration)
		VALUES (1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			media_id = excluded.media_id,
			started_at = excluded.started_at,
			duration = excluded.duration
	`,
		mediaID,
		time.Now().UnixMilli(),
		duration,
	)

	return err
}

func (s *Store) GetSync() (*model.SyncState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var syncState model.SyncState

	err := s.DB.QueryRow(`
		SELECT
			s.media_id,
			m.name,
			m.type,
			m.url,
			s.started_at,
			s.duration
		FROM sync_state s
		JOIN media m ON m.id = s.media_id
		WHERE s.id = 1
	`).Scan(
		&syncState.MediaID,
		&syncState.MediaName,
		&syncState.MediaType,
		&syncState.MediaURL,
		&syncState.StartedAt,
		&syncState.Duration,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &syncState, nil
}

func (s *Store) ClearExpiredSync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.DB.Exec(`
		DELETE FROM sync_state
		WHERE id = 1
		AND (? - started_at) >= duration * 1000
	`,
		time.Now().UnixMilli(),
	)

	return err
}

func (s *Store) Close() error {
	if s.DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	return s.DB.Close()
}