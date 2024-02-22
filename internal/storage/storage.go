package storage

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	Create(character Character) error
	List() ([]Character, error)
	Get(id string) (Character, error)
	Delete(id string) error
}

type inMemoryStorage struct {
	sync.RWMutex
	db map[string]Character
}

func (i *inMemoryStorage) Create(character Character) error {
	i.Lock()
	defer i.Unlock()

	i.db[character.ID] = character

	return nil
}

func (i *inMemoryStorage) List() ([]Character, error) {
	i.RLock()
	defer i.RUnlock()

	resp := make([]Character, 0, len(i.db))

	for _, v := range i.db {
		resp = append(resp, v)
	}

	return resp, nil
}

func (i *inMemoryStorage) Get(id string) (Character, error) {
	i.RLock()
	defer i.RUnlock()

	c, ok := i.db[id]
	if !ok {
		return Character{}, ErrNotFound
	}

	return c, nil
}

func (i *inMemoryStorage) Delete(id string) error {
	i.Lock()
	defer i.Unlock()

	if _, ok := i.db[id]; !ok {
		return ErrNotFound
	}

	delete(i.db, id)

	return nil
}

func NewInMemoryStorage() Storage {
	return &inMemoryStorage{
		RWMutex: sync.RWMutex{},
		db:      make(map[string]Character),
	}
}
