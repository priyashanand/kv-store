package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Store struct{
	data map[string]Item
	mu sync.RWMutex
}

type Item struct{
	val string
	expiresAt time.Time
}

func (s *Store) Put (key, value string, ttl int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = Item{value, time.Now().Add(time.Duration(ttl)*time.Second)}
}

func (s *Store) Get (key string) (string, error){
  s.mu.RLock()
	defer s.mu.RUnlock()
	value, err := s.data[key]
	if !err || time.Now().After(value.expiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()

		return " ", errors.New("key not found or expired")
	}
	return value.val, nil
}

func (s *Store) Delete (key string) error{
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.data[key]; !err {
		return errors.New("key not found")
	}
	delete(s.data, key)
	return nil
}

func main(){
	fmt.Println("This is my attempt at distributed systems")
	
}
