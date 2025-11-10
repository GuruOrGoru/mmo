package objects

import "maps"

import "sync"

type SharedCollection[T any] struct {
	objectsMap map[uint64]T // Mimicing database
	nextId uint64 // If for the next client
	mapMux sync.Mutex // To make it block when one client is accessing a method
}

func NewSharedCollection[T any](capacity int) *SharedCollection[T] { // Don't use variadic argument, it introduces confusion in our case
	var myLittleObjMap map[uint64]T

	if capacity <= 0 {
		myLittleObjMap = make(map[uint64]T)
	} else {
		myLittleObjMap = make(map[uint64]T, capacity)
	}

	return &SharedCollection[T]{
		objectsMap: myLittleObjMap,
		nextId: 1,
	}
}

func (s *SharedCollection[T]) Add(obj T, id ...uint64) uint64 {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()

	currId := s.nextId
	if len(id) > 0 {
		currId = id[0]
	}

	s.objectsMap[currId] = obj
	s.nextId++

	return currId
}

func (s *SharedCollection[T]) Delete(id uint64) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()

	delete(s.objectsMap,id)
}

// Loops through each elements and runs the function passed to it.
// Uses localCopy to let free the actual map for simpler methods

func (s *SharedCollection[T]) ForEach(callback func(uint64, T))  {
	s.mapMux.Lock()
	localCopy := make(map[uint64]T, len(s.objectsMap))
	maps.Copy(localCopy, s.objectsMap)
	s.mapMux.Unlock()

	// Use localCopy for callback

	for id, obj := range localCopy {
		callback(id, obj)
	}
}

func (s *SharedCollection[T]) GetObjById(id uint64) (T, bool) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()

	obj, ok := s.objectsMap[id]
	return obj, ok
}

func (s *SharedCollection[T]) ApproxLen() int {
	return len(s.objectsMap)
}
