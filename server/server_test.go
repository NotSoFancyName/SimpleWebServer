package server

import (
	"testing"
	"time"
)

func TestQueries(t *testing.T) {
	const block1 = 11509797
	bi1, err := queryBlockInfo(block1)
	if err != nil {
		t.Errorf("failed to query block %v info: %v", block1, err)
	}
	c1, err := queryTransactionsCount(block1)
	if err != nil {
		t.Errorf("failed to query block %v count: %v", block1, err)
	}
	if bi1.Transactions != c1 {
		t.Errorf("transactions count do not match: %v and %v", bi1.Transactions, c1)
	}

	const block2 = 11509797
	bi2, err := queryBlockInfo(block2)
	if err != nil {
		t.Errorf("failed to query block %v info: %v", block2, err)
	}
	c2, err := queryTransactionsCount(block1)
	if err != nil {
		t.Errorf("failed to query block %v count: %v", block1, err)
	}
	if bi2.Transactions != c2 {
		t.Errorf("transactions count do not match: %v and %v", bi2.Transactions, c2)
	}

	const block3 = 11509797
	bi3, err := queryBlockInfo(block3)
	if err != nil {
		t.Errorf("failed to query block %v info: %v", block3, err)
	}
	c3, err := queryTransactionsCount(block1)
	if err != nil {
		t.Errorf("failed to query block %v count: %v", block1, err)
	}
	if bi3.Transactions != c3 {
		t.Errorf("transactions count do not match: %v and %v", bi3.Transactions, c3)
	}
}

func TestCache(t *testing.T) {
	cache := NewCache(3, 3*time.Second)

	_, present := cache.get(777)
	if present || len(cache.blocks) != 0 {
		t.Error("The cache should be empty")
	}

	cache.put(777, &blockInfo{})
	_, present = cache.get(777)
	if !present || len(cache.blocks) != 1 {
		t.Errorf("The cache should contain one element, actual len: %v", len(cache.blocks))
	}

	cache.put(778, &blockInfo{})
	cache.put(779, &blockInfo{})
	cache.put(780, &blockInfo{})
	if len(cache.blocks) != 3 {
		t.Errorf("The cache should contain three elements, actual len: %v", len(cache.blocks))
	}

	_, present = cache.get(778)
	if !present {
		t.Error("Should be present block 778")
	}
	_, present = cache.get(779)
	if !present {
		t.Error("Should be present block 779")
	}
	_, present = cache.get(780)
	if !present {
		t.Error("Should be present block 780")
	}

	time.Sleep(4 * time.Second)
	if len(cache.blocks) != 0 {
		t.Error("The cache should be empty, the entries should be deleted")
	}
}
