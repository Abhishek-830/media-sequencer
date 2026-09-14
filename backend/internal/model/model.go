package model

type Window struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Playlist []Media `json:"playlist"`
}

type Media struct {
	ID       int    `json:"id"`
	WindowID int    `json:"window_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
}

type SyncRequest struct {
	MediaID int `json:"media_id"`
	Duration int `json:"duration"`
}

type SyncState struct {
	MediaID    int    `json:"media_id"`
	MediaName  string `json:"media_name"`
	MediaType  string `json:"media_type"`
	MediaURL   string `json:"media_url"`
	StartedAt  int64  `json:"started_at"`
	Duration   int    `json:"duration"`
}