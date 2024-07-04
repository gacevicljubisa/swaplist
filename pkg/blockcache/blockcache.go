package blockcache

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Cache struct {
	b *types.Block
	m sync.Mutex
}

func New() *Cache {
	return &Cache{}
}

func (c *Cache) Set(b *types.Block) {
	c.m.Lock()
	c.b = b
	c.m.Unlock()
}

func (c *Cache) Get(blockHash common.Hash) *types.Block {
	c.m.Lock()
	defer c.m.Unlock()
	if c.b != nil && c.b.Hash() == blockHash {
		return c.b
	}
	return nil
}
