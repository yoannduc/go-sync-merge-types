package mergetype

import (
	"sync"

	"github.com/yoannduc/mergesync/internal/entity"
)

type MergeHashMap struct {
	sync.Mutex
	m map[string]*MergeType
}

func (m *MergeHashMap) SetTypeOne(k string, v *entity.TypeOne) {
	m.Lock()
	defer m.Unlock()
	_, ok := m.m[k]
	if !ok {
		m.m[k] = &MergeType{Name: k}
	}
	m.m[k].TypeOne = v
}

func (m *MergeHashMap) SetTypeTwo(k string, v *entity.TypeTwo) {
	m.Lock()
	defer m.Unlock()
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

func NewMergeHashMap() *MergeHashMap {
	return &MergeHashMap{m: make(map[string]*MergeType)}
}
