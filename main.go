package main

import (
	"fmt"
	"hash"
	"math/rand"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

var hashfns []hash.Hash32

func init() {

	hashfns = make([]hash.Hash32, 0)
	for i := 0; i < 100; i++ {
		hashfns = append(hashfns, murmur3.New32WithSeed(rand.Uint32()))
	}

}

func murmurhash(key string, size int32, hashFnIdx int) int32 {
	hashfns[hashFnIdx].Write([]byte(key))
	result := hashfns[hashFnIdx].Sum32() % uint32(size)
	hashfns[hashFnIdx].Reset()
	return int32(result)
}

type BloomFilter struct {
	filter []uint8
	size   int32
}

func NewBloomFilter(size int32) *BloomFilter {
	return &BloomFilter{
		filter: make([]uint8, size),
		size:   size,
	}
}

func (b *BloomFilter) Add(key string, numHashfns int) {
	for i := 0; i < numHashfns; i++ {
		idx := murmurhash(key, b.size, i)
		aIdx := idx / 8
		bIdx := idx % 8
		b.filter[aIdx] = b.filter[aIdx] | (1 << bIdx)
	}
}

func (b *BloomFilter) Print() {
	fmt.Println(b.filter)
}

func (b *BloomFilter) Exists(key string, numHashfns int) (string, int32, bool) {
	for i := 0; i < numHashfns; i++ {
		idx := murmurhash(key, b.size, i)
		aIdx := idx / 8
		bIdx := idx % 8
		exist := b.filter[aIdx]&(1<<bIdx) > 0
		if !exist {
			return key, idx, false
		}
	}
	return key, 0, true
}

func main() {
	dataset := make([]string, 0)
	dataset_exists := make(map[string]bool)
	dataset_notexists := make(map[string]bool)

	for i := 0; i < 500; i++ {
		u := uuid.New()
		dataset = append(dataset, u.String())
		dataset_exists[u.String()] = true
	}

	for i := 0; i < 500; i++ {
		u := uuid.New()
		dataset = append(dataset, u.String())
		dataset_notexists[u.String()] = false
	}

	for i := 1; i < len(hashfns); i++ {
		bloom := NewBloomFilter(int32(10000))

		for key, _ := range dataset_exists {
			bloom.Add(key, i)
		}

		falsePositive := 0
		for _, key := range dataset {
			_, _, exists := bloom.Exists(key, i)
			if exists {
				if _, ok := dataset_notexists[key]; ok {
					falsePositive++
				}
			}
		}
		fmt.Println((float64(falsePositive) / float64(len(dataset))))
	}
}
