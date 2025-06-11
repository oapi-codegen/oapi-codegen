package issue1932

import (
	"container/list"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type server struct {
	cache Cache[string, string]
}

func NewServer() *server {
	return &server{
		cache: NewCache[string, string](),
	}
}

var _ ServerInterface = &server{}

// Cache defines cache interface
type Cache[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
}

// cacheImpl provides Cache interface implementation.
type cacheImpl[K comparable, V any] struct {
	sync.Mutex
	items     map[K]*list.Element
	evictList *list.List
}

// NewCache returns a new Cache.
func NewCache[K comparable, V any]() Cache[K, V] {
	return &cacheImpl[K, V]{
		items:     map[K]*list.Element{},
		evictList: list.New(),
	}
}

// Set key, ttl of 0 would use cache-wide TTL
func (c *cacheImpl[K, V]) Set(key K, value V) {
	c.add(key, value)
}

func (c *cacheImpl[K, V]) add(key K, value V) {
	c.Lock()
	defer c.Unlock()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*cacheItem[K, V]).value = value
		return
	}

	// Add new item
	ent := &cacheItem[K, V]{key: key, value: value}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry
}

// Get returns the key value if it's not expired
func (c *cacheImpl[K, V]) Get(key K) (V, bool) {
	def := *new(V)
	c.Lock()
	defer c.Unlock()
	if ent, ok := c.items[key]; ok {
		return ent.Value.(*cacheItem[K, V]).value, true
	}
	return def, false
}

type cacheItem[K comparable, V any] struct {
	key   K
	value V
}

// GetParam function is used to get param make sure right param is passed
func (s *server) GetParam(c *fiber.Ctx, param string) error {
	copiedParam := strings.Clone(param)
	time.Sleep(300 * time.Millisecond)
	s.cache.Set(param, copiedParam)

	if val, ok := s.cache.Get(copiedParam); ok && val != "" {
		return c.JSON(SimpleParam{
			Message: val,
		})
	}

	return c.SendStatus(fiber.StatusNotFound)
}
