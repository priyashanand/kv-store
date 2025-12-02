package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const BucketName = "MyBucket"

type Store struct {
	// data map[string]Item
	db *bolt.DB
	mu   sync.RWMutex
}

type Item struct {
	Val       string `json:"val"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewStore() *Store{
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error{
		_, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		return err
	})

	return &Store{
		db : db,
	}
}

func decode(val []byte) Item{
	var res Item
	err := json.Unmarshal(val, &res)
	if err != nil {
		log.Printf("Error decoding value: %v", err)
	}
	return res
}

func (s *Store) Put(key, value string, ttl int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// s.data[key] = Item{value, time.Now().Add(time.Duration(ttl) * time.Second)}

	realVal := Item{
		Val : value,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	encodedRealVal, err := json.Marshal(realVal)
	if err != nil {
			fmt.Errorf("could not encode the json ", err)
	}

	s.db.Update(func(tx *bolt.Tx) (error){
		b := tx.Bucket([]byte(BucketName))
		err := b.Put([]byte(key), []byte(encodedRealVal))
		return err
	})
}

func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var encodedVal []byte
	err := s.db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return errors.New("Bucket does not exist")
		}
		encodedVal = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		return "", err
	}

	if encodedVal == nil {
		return "", errors.New("key not found")
	}
	decval := decode(encodedVal)
	// fmt.Printf("val of %s is %s", key, decval)

	if time.Now().After(decval.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		s.db.Update(func(tx *bolt.Tx) error{
			b := tx.Bucket([]byte(BucketName))
			return b.Delete([]byte(key))
		})
		s.mu.Unlock()
		s.mu.RLock()
		return " ", errors.New("key not found or expired")
	}
	return decval.Val, nil
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(BucketName))
		return b.Delete([]byte(key))
	})
	return err
}

func main() {
	fmt.Println("This is my attempt at distributed systems")

	// db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	// if err != nil {
	// 	db.Close()
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// db.Update(func(tx *bolt.Tx) (error){
	// 	_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
	// 	if err != nil {
	// 		return fmt.Errorf("create bucket: %s", err)
	// 	}
	// 	return nil
	// })

	// db.Update(func(tx *bolt.Tx) (error){
	// 	b := tx.Bucket([]byte("MyBucket"))
	// 	err := b.Put([]byte("name1"), []byte("John"))
	// 	return err
	// })

	// db.View(func(tx *bolt.Tx) error{
	// 	b := tx.Bucket([]byte("MyBucket"))
	// 	v := b.Get([]byte("name1"))
	// 	fmt.Println("The value of name1 key is :- ", string(v))
	// 	return nil
	// })


	// store := &Store{
	// 	data: make(map[string]Item),
	// }

	store := NewStore()
	defer store.db.Close()

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
