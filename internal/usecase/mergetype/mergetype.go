package mergetype

import (
	"strconv"
	"sync"

	"github.com/yoannduc/mergesync/internal/entity"
)

type MergeType struct {
	Name    string          `json:"name"`
	TypeOne *entity.TypeOne `json:"typeOne"`
	TypeTwo *entity.TypeTwo `json:"tpyeTwo"`
}

func (m *MergeType) ToSend() bool {
	if m.TypeOne != nil && m.TypeTwo != nil {
		return true
	}

	return false
}

func TypeOneChannel(wg *sync.WaitGroup, nb int) <-chan *entity.TypeOne {
	wg.Add(1)
	out := make(chan *entity.TypeOne)

	go func() {
		defer close(out)
		for i := 0; i < nb; i++ {
			out <- &entity.TypeOne{
				Name:                 strconv.Itoa(i),
				SpecificTypeOneValue: i,
			}
		}
	}()

	return out
}

func TypeTwoChannel(wg *sync.WaitGroup, nb int) <-chan *entity.TypeTwo {
	wg.Add(1)
	out := make(chan *entity.TypeTwo)

	go func() {
		defer close(out)
		for i := 0; i < nb; i++ {
			out <- &entity.TypeTwo{
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
	m := NewMergeHashMap()

	ct1 := TypeOneChannel(wg, nb)
	ct2 := TypeTwoChannel(wg, nb)

	go func() {
		defer wg.Done()
		for v := range ct1 {
			n := v.Name
			m.SetTypeOne(n, v)

			if mt, ok := m.ToSend(n); ok {
				out <- mt
			}
		}
	}()
	go func() {
		defer wg.Done()
		for v := range ct2 {
			n := v.Name
			m.SetTypeTwo(n, v)

			if mt, ok := m.ToSend(n); ok {
				out <- mt
			}
		}
	}()

	go func() {
		wg.Wait()
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
