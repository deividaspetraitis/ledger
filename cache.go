package ledger

import (
	"log"
	"time"

	"github.com/patrickmn/go-cache"
)

type ReadFallbackFunc[Item any] func(id string) (Item, error)

func NewWithCache[Item any](c Cache[Item], readFallback ReadFallbackFunc[Item]) *WithCache[Item] {
	return &WithCache[Item]{
		cache:        c,
		readFallback: readFallback,
	}
}

type WithCache[Item any] struct {
	cache        Cache[Item]
	readFallback func(id string) (Item, error)
}

func (w *WithCache[Item]) Get(id string) (Item, error) {
	item, ok := w.cache.Read(id)
	if ok {
		return item, nil
	}

	item, err := w.readFallback(id)
	if err != nil {
		return *new(Item), nil
	}

	w.cache.Set(id, item)

	return item, nil
}

type Cache[Item any] interface {
	Read(id string) (_ Item, ok bool)
	Set(id string, item Item)
}

type InMemory[Item any] struct {
	items *cache.Cache
}

const (
	defaultExpiration = 5 * time.Minute
	purgeTime         = 10 * time.Minute
)

func NewInMemory[Item any]() *InMemory[Item] {
	return &InMemory[Item]{
		items: cache.New(defaultExpiration, purgeTime),
	}
}

func (c *InMemory[Item]) Read(id string) (_ Item, ok bool) {
	wallet, ok := c.items.Get(id)
	if ok {
		log.Println("from cache")
		return wallet.(Item), true
	}
	return *new(Item), false
}

func (c *InMemory[Item]) Set(id string, item Item) {
	c.items.Set(id, item, cache.DefaultExpiration)
}
