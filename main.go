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
	value, ok := s.data[key]
	if !ok || time.Now().After(value.expiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return " ", errors.New("key not found or expired")
	}
	val := value.val
	s.mu.RUnlock()
	return val, nil
}


func (s *Store) Delete (key string) error{
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return errors.New("key not found")
	}
	delete(s.data, key)
	return nil
}

func main(){
	fmt.Println("This is my attempt at distributed systems")

	store := &Store{
		data: make(map[string]Item),
	}

	store.Put("name1", "jhon", 3)
	store.Put("name2", "alice", 5)
	store.Put("name3", "bob", 7)

	val, _ := store.Get("name1")
	fmt.Println("GET", val)

	val, _ = store.Get("name2")
	fmt.Println("GET", val)

	val, _ = store.Get("name3")
	fmt.Println("GET", val)

	time.Sleep(4* time.Second)

	val, err := store.Get("name1")
	fmt.Println("GET", val, err)

	val, err = store.Get("name3")
	fmt.Println("GET", val, err)

	err = store.Delete("name3")
	fmt.Println("DELETE", err)

	val, err = store.Get("name3")
	fmt.Println("GET", val, err)

}
