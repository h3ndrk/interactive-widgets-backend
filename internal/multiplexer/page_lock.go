package multiplexer

import (
	"sync"

	"github.com/h3ndrk/containerized-playground/internal/id"
)

type pageLock struct {
	mutex     sync.Mutex
	pageLocks map[id.PageID]*sync.Mutex
}

func (i *pageLock) lock(pageID id.PageID) {
	i.mutex.Lock()
	closingMutex, ok := i.pageLocks[pageID]
	if !ok {
		closingMutex = &sync.Mutex{}
		i.pageLocks[pageID] = closingMutex
	}
	i.mutex.Unlock()
	closingMutex.Lock()
}

func (i *pageLock) unlock(pageID id.PageID) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	closingMutex, ok := i.pageLocks[pageID]
	if !ok {
		return
	}
	closingMutex.Unlock()
	delete(i.pageLocks, pageID)
}
