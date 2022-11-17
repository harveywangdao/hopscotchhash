package main

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"log"
)

type Entry struct {
	Key    int
	Value  interface{}
	Offset int
}

type RobinHoodHashTable struct {
	table      []*Entry
	size       int
	loadFactor float64
}

func NewRobinHoodHashTable(size int, loadFactor float64) *RobinHoodHashTable {
	arr := make([]*Entry, size)
	return &RobinHoodHashTable{
		table:      arr,
		size:       0,
		loadFactor: loadFactor,
	}
}

func (r *RobinHoodHashTable) Set(key int, value interface{}) {
	r.set(key, value)
}

func (r *RobinHoodHashTable) set(key int, value interface{}) {
	entry := &Entry{Key: key, Value: value}
	idx := r.hasher(key)

	for r.table[idx] != nil {
		if entry.Offset > r.table[idx].Offset {
			temp := r.table[idx]
			r.table[idx] = entry
			entry = temp
			idx = r.increment(idx)
			entry.Offset++
		} else if entry.Offset == r.table[idx].Offset {
			if entry.Key == r.table[idx].Key {
				r.table[idx].Value = entry.Value
				return
			} else {
				idx = r.increment(idx)
				entry.Offset++
			}
		} else {
			idx = r.increment(idx)
			entry.Offset++
		}
	}

	r.table[idx] = entry
	r.size++

	if float64(r.size) >= float64(len(r.table))*r.loadFactor {
		r.rehash(2 * len(r.table))
	}
}

func (r *RobinHoodHashTable) Print() {
	for _, e := range r.table {
		if e != nil {
			log.Println(*e)
		} else {
			log.Println("nil")
		}
	}
}

func (r *RobinHoodHashTable) rehash(newCap int) {
	oldTable := r.table
	newTable := make([]*Entry, newCap)
	r.size = 0
	r.table = newTable
	//log.Printf("rehash start, cap: %d", newCap)
	for _, e := range oldTable {
		if e != nil {
			r.set(e.Key, e.Value)
		}
	}
}

func (r *RobinHoodHashTable) increment(idx int) int {
	idx++
	if idx >= len(r.table) {
		return 0
	}
	return idx
}

func (r *RobinHoodHashTable) decrement(idx int) int {
	idx--
	if idx < 0 {
		return len(r.table) - 1
	}
	return idx
}

func (r *RobinHoodHashTable) Get(key int) (interface{}, bool) {
	offset := 0
	idx := r.hasher(key)

	for r.table[idx] != nil {
		if offset > r.table[idx].Offset {
			return nil, false
		} else if offset == r.table[idx].Offset {
			if r.table[idx].Key == key {
				return r.table[idx].Value, true
			} else {
				offset++
				idx = r.increment(idx)
			}
		} else {
			offset++
			idx = r.increment(idx)
		}
	}

	return nil, false
}

func (r *RobinHoodHashTable) Delete(key int) {
	offset := 0
	idx := r.hasher(key)

	for r.table[idx] != nil {
		if offset > r.table[idx].Offset {
			return
		} else if offset == r.table[idx].Offset {
			if r.table[idx].Key == key {
				r.table[idx] = nil
				r.size--
				idx = r.increment(idx)
				for r.table[idx] != nil && r.table[idx].Offset > 0 {
					temp := r.table[idx]
					temp.Offset--
					r.table[r.decrement(idx)] = temp
					r.table[idx] = nil
					idx = r.increment(idx)
				}
				return
			} else {
				offset++
				idx = r.increment(idx)
			}
		} else {
			offset++
			idx = r.increment(idx)
		}
	}
}

func (r *RobinHoodHashTable) Range(f func(key int, value interface{}) bool) {
	for i := 0; i < len(r.table); i++ {
		if r.table[i] != nil {
			ok := f(r.table[i].Key, r.table[i].Value)
			if !ok {
				return
			}
		}
	}
}

func (r *RobinHoodHashTable) hasher(key int) int {
	return key % len(r.table)
}

func (r *RobinHoodHashTable) hasher2(key int) int {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, uint64(key)); err != nil {
		log.Fatal(err)
	}
	f := fnv.New64()
	if _, err := f.Write(buf.Bytes()); err != nil {
		log.Fatal(err)
	}
	total := len(r.table)
	hashCode := f.Sum64() % uint64(total)
	//log.Printf("key: %d, hashcode: %d", key, int(hashCode))
	return int(hashCode)
}
