package service

import (
	"errors"
	"strings"

	"media-sequencer/backend/internal/model"
	"media-sequencer/backend/internal/store"
)

type Service struct {
	store *store.Store
}

func New(s *store.Store) *Service {
	return &Service{
		store: s,
	}
}

func (s *Service) GetWindows() ([]model.Window, error) {
	return s.store.GetWindows()
}

func (s *Service) AddMedia(
	windowID int,
	name string,
	mediaType string,
	url string,
	duration int,
) (*model.Media, error) {

	if windowID <= 0 {
		return nil, errors.New("invalid window id")
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, errors.New("media name is required")
	}

	mediaType = strings.ToLower(strings.TrimSpace(mediaType))

	if mediaType != "image" &&
		mediaType != "video" &&
		mediaType != "blank" {
		return nil, errors.New("media type must be image, video, or blank")
	}

	if duration <= 0 {
		return nil, errors.New("duration must be greater than zero")
	}

	if mediaType != "blank" {
		url = strings.TrimSpace(url)

		if url == "" {
			return nil, errors.New("media URL is required")
		}
	} else {
		url = ""
	}

	return s.store.AddMedia(
		windowID,
		name,
		mediaType,
		url,
		duration,
	)
}

func (s *Service) SyncMedia(
	mediaID int,
	duration int,
) (*model.SyncState, error) {

	if mediaID <= 0 {
		return nil, errors.New("invalid media id")
	}

	if duration <= 0 {
		return nil, errors.New("sync duration must be greater than zero")
	}

	media, err := s.store.GetMediaByID(mediaID)
	if err != nil {
		return nil, errors.New("media not found")
	}

	if err := s.store.SetSync(mediaID, duration); err != nil {
		return nil, err
	}

	syncState, err := s.store.GetSync()
	if err != nil {
		return nil, err
	}

	if syncState == nil {
		return nil, errors.New("sync state could not be created")
	}

	_ = media

	return syncState, nil
}

func (s *Service) GetSync() (*model.SyncState, error) {
	if err := s.store.ClearExpiredSync(); err != nil {
		return nil, err
	}

	return s.store.GetSync()
}