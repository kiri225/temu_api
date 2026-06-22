package main

import (
	"os"
	"sync"

	"github.com/goccy/go-json"
)

type apiSampleEntry struct {
	Body       json.RawMessage `json:"body,omitempty"`
	ParamNotes map[string]any  `json:"paramNotes,omitempty"`
}

type apiSamplesData struct {
	ByID   map[string]apiSampleEntry `json:"byId"`
	ByType map[string]apiSampleEntry `json:"byType"`
}

type sampleStore struct {
	path string
	mu   sync.Mutex
	data apiSamplesData
}

func newSampleStore(path string) (*sampleStore, error) {
	s := &sampleStore{
		path: path,
		data: apiSamplesData{
			ByID:   map[string]apiSampleEntry{},
			ByType: map[string]apiSampleEntry{},
		},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *sampleStore) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.save()
		}
		return err
	}
	var data apiSamplesData
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if data.ByID == nil {
		data.ByID = map[string]apiSampleEntry{}
	}
	if data.ByType == nil {
		data.ByType = map[string]apiSampleEntry{}
	}
	s.data = data
	return nil
}

func (s *sampleStore) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.path, b, 0644)
}

func (s *sampleStore) snapshot() apiSamplesData {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cloneData()
}

func (s *sampleStore) cloneData() apiSamplesData {
	byID := make(map[string]apiSampleEntry, len(s.data.ByID))
	for k, v := range s.data.ByID {
		byID[k] = v
	}
	byType := make(map[string]apiSampleEntry, len(s.data.ByType))
	for k, v := range s.data.ByType {
		byType[k] = v
	}
	return apiSamplesData{ByID: byID, ByType: byType}
}

func (s *sampleStore) entry(id, typ string) (apiSampleEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		if e, ok := s.data.ByID[id]; ok {
			return e, true
		}
	}
	if typ != "" {
		if e, ok := s.data.ByType[typ]; ok {
			return e, true
		}
	}
	return apiSampleEntry{}, false
}

type sampleUpdateRequest struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Body       json.RawMessage `json:"body"`
	ParamNotes map[string]any  `json:"paramNotes"`
	Clear      bool            `json:"clear"`
}

func (s *sampleStore) apply(req sampleUpdateRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := apiSampleEntry{
		Body:       req.Body,
		ParamNotes: req.ParamNotes,
	}

	if req.Clear {
		if req.ID != "" {
			delete(s.data.ByID, req.ID)
		}
		if req.Type != "" {
			delete(s.data.ByType, req.Type)
		}
		return s.save()
	}

	if len(req.Body) == 0 && len(req.ParamNotes) == 0 {
		return nil
	}

	if req.ID != "" {
		s.data.ByID[req.ID] = entry
	}
	if req.Type != "" {
		s.data.ByType[req.Type] = entry
	}
	return s.save()
}

func applySampleOverrides(apis []apiEntry, store *sampleStore) []apiEntry {
	out := make([]apiEntry, len(apis))
	for i, api := range apis {
		out[i] = api
		e, ok := store.entry(api.ID, api.Type)
		if !ok {
			continue
		}
		if len(e.Body) > 0 && string(e.Body) != "null" {
			out[i].SampleBody = string(e.Body)
		}
		if len(e.ParamNotes) > 0 {
			out[i].ParamNotes = e.ParamNotes
		}
	}
	return out
}
