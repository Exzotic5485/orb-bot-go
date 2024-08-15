package store

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Store struct {
	data     map[string]int
	lock     sync.RWMutex
	filePath string
}

func NewStore(filePath string) *Store {
	s := &Store{
		data:     make(map[string]int),
		filePath: filePath,
	}

	s.load()

	return s
}

func (s *Store) Get(key string) int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data[key]
}

func (s *Store) Set(key string, value int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = value

	s.save()
}

func (s *Store) Increment(key string, amount int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] += amount

	s.save()
}

func (s *Store) save() {
	fileContents, err := json.Marshal(s.data)

	if err != nil {
		log.Fatalf("Failed to marshal store data '%v': %v", s.filePath, err)
	}

	err = os.WriteFile(s.filePath, fileContents, 0666)

	if err != nil {
		log.Fatalf("Failed to write store to disk '%v': %v", s.filePath, err)
	}
}

func (s *Store) load() {
	s.lock.Lock()
	defer s.lock.Unlock()

	fileContents, err := os.ReadFile(s.filePath)

	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		log.Fatalf("Failed to read store '%v': %v", s.filePath, err)
	}

	err = json.Unmarshal(fileContents, &s.data)

	if err != nil {
		log.Fatalf("Failed to parse store '%v': %v", s.filePath, err)
	}
}
