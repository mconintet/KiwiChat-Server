package main

import "sync"

type _uidConnIdMap struct {
	uidConnId map[int64]uint64
	connIdUid map[uint64]int64
	mu        sync.Mutex
}

func newUidConnIdMap() *_uidConnIdMap {
	m := &_uidConnIdMap{}
	m.uidConnId = make(map[int64]uint64)
	m.connIdUid = make(map[uint64]int64)
	return m
}

func (m *_uidConnIdMap) add(uid int64, connId uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.uidConnId[uid] = connId
	m.connIdUid[connId] = uid
}

func (m *_uidConnIdMap) getConnId(uid int64) (connId uint64, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connId, ok = m.uidConnId[uid]
	return
}

func (m *_uidConnIdMap) getUid(connId uint64) (uid int64, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid, ok = m.connIdUid[connId]
	return
}

func (m *_uidConnIdMap) delConnId(connId uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid, ok := m.connIdUid[connId]
	if ok {
		delete(m.uidConnId, uid)
	}
	delete(m.connIdUid, connId)
}

func (m *_uidConnIdMap) delUid(uid int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connId, ok := m.uidConnId[uid]
	if ok {
		delete(m.connIdUid, connId)
	}
	delete(m.uidConnId, uid)
}

var uidConnIdMap = newUidConnIdMap()
