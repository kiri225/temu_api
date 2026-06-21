package main

import (
	"os"
	"sync"

	"github.com/goccy/go-json"
)

type unavailableEntry struct {
	Note string `json:"note,omitempty"`
}

type unavailableData struct {
	ByID   map[string]unavailableEntry `json:"byId"`
	ByType map[string]unavailableEntry `json:"byType"`
}

type unavailableStore struct {
	path string
	mu   sync.Mutex
	data unavailableData
}

func newUnavailableStore(path string) (*unavailableStore, error) {
	s := &unavailableStore{
		path: path,
		data: unavailableData{
			ByID:   map[string]unavailableEntry{},
			ByType: map[string]unavailableEntry{},
		},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *unavailableStore) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.save()
		}
		return err
	}
	var data unavailableData
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if data.ByID == nil {
		data.ByID = map[string]unavailableEntry{}
	}
	if data.ByType == nil {
		data.ByType = map[string]unavailableEntry{}
	}
	s.data = data
	return nil
}

func (s *unavailableStore) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.path, b, 0644)
}

func (s *unavailableStore) snapshot() unavailableData {
	s.mu.Lock()
	defer s.mu.Unlock()
	byID := make(map[string]unavailableEntry, len(s.data.ByID))
	for k, v := range s.data.ByID {
		byID[k] = v
	}
	byType := make(map[string]unavailableEntry, len(s.data.ByType))
	for k, v := range s.data.ByType {
		byType[k] = v
	}
	return unavailableData{ByID: byID, ByType: byType}
}

func (s *unavailableStore) isUnavailable(id, typ string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		if _, ok := s.data.ByID[id]; ok {
			return true
		}
	}
	if typ != "" {
		if _, ok := s.data.ByType[typ]; ok {
			return true
		}
	}
	return false
}

func (s *unavailableStore) note(id, typ string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		if e, ok := s.data.ByID[id]; ok && e.Note != "" {
			return e.Note
		}
	}
	if typ != "" {
		if e, ok := s.data.ByType[typ]; ok && e.Note != "" {
			return e.Note
		}
	}
	return ""
}

type unavailableUpdateRequest struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Unavailable bool   `json:"unavailable"`
	Note        string `json:"note"`
}

func (s *unavailableStore) apply(req unavailableUpdateRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.ID != "" {
		if req.Unavailable {
			s.data.ByID[req.ID] = unavailableEntry{Note: req.Note}
		} else {
			delete(s.data.ByID, req.ID)
		}
	}
	if req.Type != "" {
		if req.Unavailable {
			s.data.ByType[req.Type] = unavailableEntry{Note: req.Note}
		} else {
			delete(s.data.ByType, req.Type)
		}
	}
	return s.save()
}

func applyUnavailableFlags(apis []apiEntry, store *unavailableStore) []apiEntry {
	out := make([]apiEntry, len(apis))
	for i, api := range apis {
		out[i] = api
		if store.isUnavailable(api.ID, api.Type) {
			out[i].Unavailable = true
			out[i].UnavailableNote = store.note(api.ID, api.Type)
		}
	}
	return out
}
