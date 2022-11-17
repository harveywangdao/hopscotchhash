package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type HashMap interface {
	Print()
	Range(f func(key int, value interface{}) bool)
	Get(key int) (interface{}, bool)
	Delete(key int)
	Set(key int, value interface{})
}

// 冲突测试
func do1(h HashMap) {
	keys := []int{1, 2, 3, 4, 5, 3 + 16}
	values := []int{1, 2, 3, 4, 5, 6}
	for i, key := range keys {
		h.Set(key, values[i])
	}
	h.Print()

	for i, key := range keys {
		v, ok := h.Get(key)
		if !ok {
			log.Fatalf("key: %d not existed", key)
		}
		if value, ok2 := v.(int); !ok2 || value != values[i] {
			log.Fatalf("key: %d, value: %v error", key, v)
		}
	}
}

// 相同key
func do2(h HashMap) {
	key := 4
	h.Set(key, 1)
	h.Set(key, 2)
	h.Set(key, 3)
	h.Print()
	v, ok := h.Get(key)
	if !ok {
		log.Fatalf("key %d not existed", key)
	}
	if value, ok2 := v.(int); !ok2 || value != 3 {
		log.Fatalf("key %d, value: %v error", key, v)
	}
}

// 删除key
func do3(h HashMap) {
	n := 5
	for i := 0; i < n; i++ {
		h.Set(i, i)
	}
	h.Print()
	fmt.Println()

	delkey := 3
	h.Delete(delkey)
	h.Print()

	_, ok := h.Get(delkey)
	if ok {
		log.Fatal("key should not existed")
	}
}

// 递增key
func do4(h HashMap) {
	n := 100000
	start := time.Now()
	for i := 0; i < n; i++ {
		h.Set(i, i)
	}
	log.Println("set cost:", time.Now().Sub(start))

	start = time.Now()
	for i := 0; i < n; i++ {
		v, ok := h.Get(i)
		if !ok {
			log.Fatalf("key %d not existed", i)
		}

		//log.Println(i, v)

		if value, ok2 := v.(int); !ok2 || value != i {
			log.Fatalf("key %d, value: %v error", i, v)
		}
	}
	log.Println("get cost:", time.Now().Sub(start))
}

// 随机key
func do5(h HashMap) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	var keys, values []int
	n := 100000
	for i := 0; i < n; i++ {
		keys = append(keys, r.Int())
		values = append(values, r.Int())
	}

	start := time.Now()
	for i, key := range keys {
		h.Set(key, values[i])
	}
	log.Println("set cost:", time.Now().Sub(start))

	start = time.Now()
	for i, key := range keys {
		v, ok := h.Get(key)
		if !ok {
			log.Fatalf("key %d not existed", key)
		}

		//log.Println(key, v)

		if value, ok2 := v.(int); !ok2 || value != values[i] {
			log.Fatalf("key %d, value: %v error", key, v)
		}
	}
	log.Println("get cost:", time.Now().Sub(start))
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//h := NewHopScotchHashTable(16, 4)
	h := NewRobinHoodHashTable(16, 0.5)
	do1(h)
}
