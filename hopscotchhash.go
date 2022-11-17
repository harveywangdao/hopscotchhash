package main

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"log"
)

type Elem struct {
	Key   int
	Value interface{}
}

type HashEntry struct {
	Dist uint64
	Elem *Elem
}

type HopScotchHashTable struct {
	arr     []*HashEntry
	maxDist int
	sz      int
}

func NewHopScotchHashTable(size, maxDist int) *HopScotchHashTable {
	arr := make([]*HashEntry, size)
	for i := 0; i < len(arr); i++ {
		arr[i] = &HashEntry{}
	}

	if maxDist > 64 {
		maxDist = 64
	}

	return &HopScotchHashTable{
		arr:     arr,
		maxDist: maxDist,
	}
}

func (h *HopScotchHashTable) Print() {
	for i := 0; i < len(h.arr); i++ {
		if h.arr[i].Elem != nil {
			log.Println(h.arr[i].Dist, *h.arr[i].Elem)
		} else {
			log.Println(h.arr[i].Dist, "nil")
		}
	}
}

func (h *HopScotchHashTable) Range(f func(key int, value interface{}) bool) {
	for i := 0; i < len(h.arr); i++ {
		if h.arr[i].Elem != nil {
			ok := f(h.arr[i].Elem.Key, h.arr[i].Elem.Value)
			if !ok {
				return
			}
		}
	}
}

func (h *HopScotchHashTable) findPos(key int) int {
	startPos := h.hasher(key)
	for i := 0; i < h.maxDist; i++ {
		if (h.arr[startPos].Dist>>i)%2 == 1 {
			pos := h.incr(startPos, h.maxDist-1-i)
			if h.arr[pos].Elem.Key == key {
				return pos
			}
		}
	}
	return -1
}

func (h *HopScotchHashTable) Get(key int) (interface{}, bool) {
	pos := h.findPos(key)
	if pos == -1 {
		return nil, false
	}
	return h.arr[pos].Elem.Value, true
}

func (h *HopScotchHashTable) Delete(key int) {
	pos := h.findPos(key)
	if pos != -1 {
		startPos := h.hasher(key)
		h.arr[pos].Elem = nil
		h.arr[startPos].Dist -= (1 << h.distShift(startPos, pos))
	}
}

func (h *HopScotchHashTable) Set(key int, value interface{}) {
	pos := h.findPos(key)
	if pos != -1 {
		h.arr[pos].Elem.Value = value
		return
	}
	h.set(key, value)
}

func (h *HopScotchHashTable) set(key int, value interface{}) {
	for {
		if h.sz >= len(h.arr) {
			h.rehash()
		}

		pos := h.hasher(key)
		startPos := pos

		for h.arr[pos].Elem != nil {
			pos = h.incr(pos, 1)
		}

		if h.matchDist(startPos, pos) {
			h.arr[pos].Elem = &Elem{Key: key, Value: value}
			h.arr[startPos].Dist += (1 << h.distShift(startPos, pos)) // 1 << (h.maxDist - 1 + 领主位置 - 领子位置)
			h.sz++
			//log.Println("insert1:", key, value)
			return
		}

		for {
			isNotDist := false

			for i := h.maxDist - 1; i > 0; i-- {
				for j := h.maxDist - 1; j > h.maxDist-1-i; j-- {
					tmpStartPos := h.decr(pos, i)
					if (h.arr[tmpStartPos].Dist>>j)%2 == 1 {
						tmpPos := h.incr(tmpStartPos, h.maxDist-1-j)
						tmp := h.arr[tmpPos]
						h.arr[pos].Elem = tmp.Elem
						tmp.Elem = nil
						// 领主位置: pos-i
						// 旧位置: pos-i+h.maxDist-1-j
						// 新位置: pos
						// 从领域摘除,再重新设置新位置
						h.arr[tmpStartPos].Dist = h.arr[tmpStartPos].Dist - (1 << j) + (1 << h.distShift(tmpStartPos, pos))

						// pos新位置,相当于pos向上移动
						pos = tmpPos

						if h.matchDist(startPos, pos) {
							h.arr[pos].Elem = &Elem{Key: key, Value: value}
							h.arr[startPos].Dist += (1 << h.distShift(startPos, pos))
							h.sz++
							//log.Println("insert2:", key, value)
							return
						} else {
							isNotDist = true
							break
						}
					}
				}

				if isNotDist {
					break
				}
			}

			if !isNotDist {
				break
			}
		}

		h.rehash()
	}
}

func (h *HopScotchHashTable) incr(pos, step int) int {
	n := pos + step
	if n >= len(h.arr) {
		return n - len(h.arr)
	}
	return n
}

func (h *HopScotchHashTable) decr(pos, step int) int {
	n := pos - step
	if n < 0 {
		return n + len(h.arr)
	}
	return n
}

func (h *HopScotchHashTable) matchDist(startPos, pos int) bool {
	if pos >= startPos {
		return pos <= startPos+h.maxDist-1
	}
	return pos+len(h.arr) <= startPos+h.maxDist-1
}

func (h *HopScotchHashTable) distShift(startPos, pos int) int {
	if pos >= startPos {
		return h.maxDist - 1 + startPos - pos
	}
	return h.maxDist - 1 + startPos - (pos + len(h.arr))
}

func (h *HopScotchHashTable) rehash() {
	oldArr := h.arr
	newArr := make([]*HashEntry, 2*len(oldArr))
	for i := 0; i < len(newArr); i++ {
		newArr[i] = &HashEntry{}
	}
	h.sz = 0
	h.arr = newArr

	//log.Printf("rehash start, cap: %d", len(h.arr))
	for i := 0; i < len(oldArr); i++ {
		if elem := oldArr[i].Elem; elem != nil {
			h.set(elem.Key, elem.Value)
		}
	}
	//log.Printf("rehash end, cap: %d", len(h.arr))
}

func (h *HopScotchHashTable) hasher(key int) int {
	return key % len(h.arr)
}

func (h *HopScotchHashTable) hasher2(key int) int {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, uint64(key)); err != nil {
		log.Fatal(err)
	}
	f := fnv.New64()
	if _, err := f.Write(buf.Bytes()); err != nil {
		log.Fatal(err)
	}
	total := len(h.arr)
	hashCode := f.Sum64() % uint64(total)
	//log.Printf("key: %d, hashcode: %d", key, int(hashCode))

	return int(hashCode)
}
