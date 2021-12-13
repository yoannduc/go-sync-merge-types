package main

import (
	"log"
	"strconv"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

type TypeOne struct {
	Name                 string `json:"name"`
	SpecificTypeOneValue int    `json:"typeOneValue"`
}

type TypeTwo struct {
	Name                 string `json:"name"`
	SpecificTypeTwoValue int    `json:"typeTwoValue"`
}

type MergeType struct {
	Name    string   `json:"name"`
	TypeOne *TypeOne `json:"typeOne"`
	TypeTwo *TypeTwo `json:"tpyeTwo"`
}

func (m *MergeType) ToSend() bool {
	if m.TypeOne != nil && m.TypeTwo != nil {
		return true
	}

	return false
}

type MergeHashMap struct {
	sync.Mutex
	m map[string]*MergeType
}

func (m *MergeHashMap) SetTypeOne(k string, v *TypeOne) {
	m.Lock()
	defer m.Unlock()
	log.Println("Entering")
	_, ok := m.m[k]
	if !ok {
		m.m[k] = &MergeType{Name: k}
	}
	m.m[k].TypeOne = v
}

func (m *MergeHashMap) SetTypeTwo(k string, v *TypeTwo) {
	m.Lock()
	defer m.Unlock()
	log.Println("Entering")
	_, ok := m.m[k]
	if !ok {
		m.m[k] = &MergeType{Name: k}
	}
	m.m[k].TypeTwo = v
}

func (m *MergeHashMap) Purge() {
	m.m = nil
}

func (m *MergeHashMap) ToSend(k string) (*MergeType, bool) {
	m.Lock()
	defer m.Unlock()
	v, ok := m.m[k]
	if ok && v.ToSend() {
		delete(m.m, k)
		return v, true
	}

	return nil, false
}

func TypeOneChannel(wg *sync.WaitGroup, nb int) <-chan *TypeOne {
	wg.Add(1)
	out := make(chan *TypeOne)

	go func() {
		defer close(out)
		for i := 0; i < nb; i++ {
			out <- &TypeOne{
				Name:                 strconv.Itoa(i),
				SpecificTypeOneValue: i,
			}
		}
	}()

	return out
}

func TypeTwoChannel(wg *sync.WaitGroup, nb int) <-chan *TypeTwo {
	wg.Add(1)
	out := make(chan *TypeTwo)

	go func() {
		defer close(out)
		for i := 0; i < nb; i++ {
			log.Printf("i2 | %T | %v\n", i, i)
			out <- &TypeTwo{
				Name:                 strconv.Itoa(i),
				SpecificTypeTwoValue: i,
			}
		}
	}()

	return out
}

func MergeChan(nb int) <-chan *MergeType {
	out := make(chan *MergeType)
	wg := new(sync.WaitGroup)
	m := &MergeHashMap{m: make(map[string]*MergeType)}

	ct1 := TypeOneChannel(wg, nb)
	ct2 := TypeTwoChannel(wg, nb)

	go func() {
		for v := range ct1 {
			log.Printf("v1 | %T | %v\n", v, v)
			n := v.Name
			m.SetTypeOne(n, v)

			if mt, ok := m.ToSend(n); ok {
				out <- mt
			}
		}
		wg.Done()
	}()
	go func() {
		for v := range ct2 {
			log.Printf("v2 | %T | %v\n", v, v)
			n := v.Name
			m.SetTypeTwo(n, v)

			if mt, ok := m.ToSend(n); ok {
				out <- mt
			}
		}
		wg.Done()
	}()

	go func() {
		// log.Printf("wg | %T | %v\n", wg, wg)
		wg.Wait()
		// if j, err := jsoniter.MarshalToString(m); err == nil {
		// 	log.Printf("m | %T | %v\n", j, j)
		// } else {
		// 	log.Printf("m | %T | %v\n", m, m)
		// }
		m.Purge()
		close(out)
	}()

	return out
}

func MergedList(nb int) []*MergeType {
	mc := MergeChan(10)

	var o []*MergeType
	for v := range mc {
		o = append(o, v)
	}

	return o
}

func main() {
	o := MergedList(10)

	if j, err := jsoniter.MarshalToString(o); err == nil {
		log.Printf("o | %T | %v\n", j, j)
	} else {
		log.Printf("o | %T | %v\n", o, o)
	}

	log.Printf("len(o) | %T | %v\n", len(o), len(o))
}
