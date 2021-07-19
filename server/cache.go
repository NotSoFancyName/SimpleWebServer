package server

import (
	"sync"
	"time"
)

type cache struct {
	blocks map[int]*cacheEntry
	usage  []int
	mtx    sync.Mutex

	maxSize int
	expTime time.Duration
}

type cacheEntry struct{
	block *blockInfo
	timer *time.Timer
}

func NewCache(size int, expiration time.Duration) *cache {
	blocks := make(map[int]*cacheEntry)
	return &cache{
		usage:   []int{},
		blocks:  blocks,
		maxSize: size,
		expTime: expiration,
	}
}

func (c *cache) put(blockNum int, b *blockInfo) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	entry, present := c.blocks[blockNum]
	if present {
		entry.timer.Stop()
		for idx, b := range c.usage {
			if b == blockNum {
				c.usage = append(c.usage[:idx], c.usage[idx+1:]...)
				break
			}
		}
	}

	timer := time.AfterFunc(c.expTime, func() {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		_, present := c.blocks[blockNum]
		if present {
			delete(c.blocks, blockNum)
			for idx, b := range c.usage {
				if b == blockNum {
					c.usage = append(c.usage[:idx], c.usage[idx+1:]...)
					break
				}
			}
		}
	})

	c.usage = append(c.usage, blockNum)
	c.blocks[blockNum] = &cacheEntry{
		block: b,
		timer: timer,
	}

	if c.maxSize < len(c.blocks) {
		delete(c.blocks, c.usage[0])
		c.usage = c.usage[1:]
	}
}

func (c *cache) get(blockNum int) (block *blockInfo, present bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	entry, present := c.blocks[blockNum]
	if !present {
		return
	}
	block = entry.block

	entry.timer.Stop()
	timer := time.AfterFunc(c.expTime, func() {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		_, present := c.blocks[blockNum]
		if present {
			delete(c.blocks, blockNum)
			for idx, b := range c.usage {
				if b == blockNum {
					c.usage = append(c.usage[:idx], c.usage[idx+1:]...)
					break
				}
			}
		}
	})
	c.blocks[blockNum] = &cacheEntry{
		block: entry.block,
		timer: timer,
	}

	for idx, b := range c.usage {
		if b == blockNum {
			c.usage = append(c.usage[:idx], c.usage[idx+1:]...)
			c.usage = append(c.usage, blockNum)
			break
		}
	}
	return
}
