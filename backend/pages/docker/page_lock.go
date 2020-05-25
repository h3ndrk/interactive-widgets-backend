package docker

import (
	"sync"

	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type PageLock struct {
	mutex     sync.Mutex
	pageLocks map[pages.PageID]*sync.Mutex
}

func (i *PageLock) Lock(pageID pages.PageID) {
	i.mutex.Lock()
	closingMutex, ok := i.pageLocks[pageID]
	if !ok {
		closingMutex = &sync.Mutex{}
		i.pageLocks[pageID] = closingMutex
	}
	i.mutex.Unlock()
	closingMutex.Lock()
}

func (i *PageLock) Unlock(pageID pages.PageID) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	closingMutex, ok := i.pageLocks[pageID]
	if !ok {
		return
	}
	closingMutex.Unlock()
	delete(i.pageLocks, pageID)
}
