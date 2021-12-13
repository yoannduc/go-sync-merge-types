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
	sync.RWMutex
	m map[string]*MergeType
}

func (m *MergeHashMap) Load(k string) (*MergeType, bool) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *MergeHashMap) Store(k string, v *MergeType) {
	m.Lock()
	defer m.Unlock()
	m.m[k] = v
}

func (m *MergeHashMap) Delete(k string) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

func (m *MergeHashMap) Purge() {
	m.m = nil
}

func (m *MergeHashMap) ToSend(k string) bool {
	v, ok := m.Load(k)
	if !ok {
		return false
	}

	if v.TypeOne != nil && v.TypeTwo != nil {
		m.Delete(k)
		return true
	}

	return false
}

type MergeHashMapBis struct {
	sync.Mutex
	m map[string]*MergeType
}

func (m *MergeHashMapBis) Load(k string) (*MergeType, bool) {
	m.Lock()
	defer m.Unlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *MergeHashMapBis) Store(k string, v *MergeType) {
	m.Lock()
	defer m.Unlock()
	m.m[k] = v
}

func (m *MergeHashMapBis) Delete(k string) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

func (m *MergeHashMapBis) Purge() {
	m.m = nil
}

type MergeHashMapFour struct {
	sync.Mutex
	m map[string]*MergeType
}

func (m *MergeHashMapFour) SetTypeOne(k string, v *TypeOne) {
	m.Lock()
	defer m.Unlock()
	log.Println("Entering")
	_, ok := m.m[k]
	if !ok {
		m.m[k] = &MergeType{Name: k}
	}
	m.m[k].TypeOne = v
}

func (m *MergeHashMapFour) SetTypeTwo(k string, v *TypeTwo) {
	m.Lock()
	defer m.Unlock()
	log.Println("Entering")
	_, ok := m.m[k]
	if !ok {
		m.m[k] = &MergeType{Name: k}
	}
	m.m[k].TypeTwo = v
}

func (m *MergeHashMapFour) Purge() {
	m.m = nil
}

func (m *MergeHashMapFour) ToSend(k string) (*MergeType, bool) {
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
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeOne = v

			log.Printf("mt1 | %T | %v\n", mt, mt)

			if mt.ToSend() {
				out <- mt
				m.Delete(n)
			} else {
				m.Store(n, mt)
			}
		}
		wg.Done()
	}()
	go func() {
		for v := range ct2 {
			log.Printf("v2 | %T | %v\n", v, v)
			n := v.Name
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeTwo = v
			log.Printf("mt2 | %T | %v\n", mt, mt)

			if mt.ToSend() {
				out <- mt
				m.Delete(n)
			} else {
				m.Store(n, mt)
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

func MergeChanBis(nb int) <-chan *MergeType {
	out := make(chan *MergeType)
	wg := new(sync.WaitGroup)
	m := &MergeHashMapBis{m: make(map[string]*MergeType)}

	ct1 := TypeOneChannel(wg, nb)
	ct2 := TypeTwoChannel(wg, nb)

	go func() {
		for v := range ct1 {
			log.Printf("v1 | %T | %v\n", v, v)
			n := v.Name
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeOne = v

			log.Printf("mt1 | %T | %v\n", mt, mt)

			if mt.ToSend() {
				out <- mt
				m.Delete(n)
			} else {
				m.Store(n, mt)
			}
		}
		wg.Done()
	}()
	go func() {
		for v := range ct2 {
			log.Printf("v2 | %T | %v\n", v, v)
			n := v.Name
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeTwo = v
			log.Printf("mt2 | %T | %v\n", mt, mt)

			if mt.ToSend() {
				out <- mt
				m.Delete(n)
			} else {
				m.Store(n, mt)
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

func MergedListBis(nb int) []*MergeType {
	mc := MergeChanBis(10)

	var o []*MergeType
	for v := range mc {
		o = append(o, v)
	}

	return o
}

func MergeChanTer(nb int) <-chan *MergeType {
	out := make(chan *MergeType)
	wg := new(sync.WaitGroup)
	m := &MergeHashMap{m: make(map[string]*MergeType)}

	ct1 := TypeOneChannel(wg, nb)
	ct2 := TypeTwoChannel(wg, nb)

	go func() {
		for v := range ct1 {
			log.Printf("v1 | %T | %v\n", v, v)
			n := v.Name
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeOne = v

			log.Printf("mt1 | %T | %v\n", mt, mt)

			m.Store(n, mt)

			if m.ToSend(n) {
				out <- mt
			}
		}
		wg.Done()
	}()
	go func() {
		for v := range ct2 {
			log.Printf("v2 | %T | %v\n", v, v)
			n := v.Name
			mt, ok := m.Load(n)
			if !ok {
				mt = &MergeType{
					Name: n,
				}
			}

			mt.TypeTwo = v
			log.Printf("mt2 | %T | %v\n", mt, mt)

			m.Store(n, mt)

			if m.ToSend(n) {
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

func MergedListTer(nb int) []*MergeType {
	mc := MergeChanTer(10)

	var o []*MergeType
	for v := range mc {
		o = append(o, v)
	}

	return o
}

func MergeChanFour(nb int) <-chan *MergeType {
	out := make(chan *MergeType)
	wg := new(sync.WaitGroup)
	m := &MergeHashMapFour{m: make(map[string]*MergeType)}

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

func MergedListFour(nb int) []*MergeType {
	mc := MergeChanFour(10)

	var o []*MergeType
	for v := range mc {
		o = append(o, v)
	}

	return o
}

func main() {
	// o := MergedList(10)
	// o := MergedListBis(10)
	// o := MergedListTer(10)
	o := MergedListFour(10)

	if j, err := jsoniter.MarshalToString(o); err == nil {
		log.Printf("o | %T | %v\n", j, j)
	} else {
		log.Printf("o | %T | %v\n", o, o)
	}

	log.Printf("len(o) | %T | %v\n", len(o), len(o))
}
