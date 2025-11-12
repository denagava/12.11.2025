package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type LinkSet struct {
	Links     map[string]string `json:"links"`
	Timestamp time.Time         `json:"timestamp"`
	LinksNum  int               `json:"links_num"`
}

type Storage struct {
	mu          sync.RWMutex
	linksSets   map[int]*LinkSet
	nextID      int
	storageFile string
}

func NewStorage(storageFile string) *Storage {
	storage := &Storage{
		linksSets:   make(map[int]*LinkSet),
		nextID:      1,
		storageFile: storageFile,
	}
	storage.loadFromFile()
	return storage
}

func (s *Storage) loadFromFile() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.storageFile)
	if err != nil {
		return
	}

	var fileData struct {
		LinksSets map[int]*LinkSet `json:"links_sets"`
		NextID    int              `json:"next_id"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return
	}

	s.linksSets = fileData.LinksSets
	s.nextID = fileData.NextID
}

func (s *Storage) saveToFile() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fileData := struct {
		LinksSets map[int]*LinkSet `json:"links_sets"`
		NextID    int              `json:"next_id"`
	}{
		LinksSets: s.linksSets,
		NextID:    s.nextID,
	}
	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return
	}
	tmpFile := s.storageFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return
	}
	os.Rename(tmpFile, s.storageFile)
}

func (s *Storage) SaveLinkSet(links map[string]string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	linksNum := s.nextID
	s.linksSets[linksNum] = &LinkSet{
		Links:     links,
		Timestamp: time.Now(),
		LinksNum:  linksNum,
	}
	s.nextID++
	go s.saveToFile()
	return linksNum
}

func (s *Storage) GetLinkSets(linksNums []int) map[int]*LinkSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int]*LinkSet)
	for _, num := range linksNums {
		if linkSet, exists := s.linksSets[num]; exists {
			result[num] = linkSet
		}
	}
	return result
}

func (s *Storage) GetAllData() map[int]*LinkSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int]*LinkSet)
	for k, v := range s.linksSets {
		result[k] = v
	}
	return result
}
