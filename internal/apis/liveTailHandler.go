package apis

import (
	"net/http"
	"sync"

	"github.com/5733ccV10/go-log/internal/storage"
)

const viewerBufferCapacity = 100

type liveTailRegistry struct {
	Registry map[string][]chan storage.Entry
	mu       sync.Mutex
}

func liveTailHandler(w http.ResponseWriter, r *http.Request) {

}

func (ltr *liveTailRegistry) Subscribe(serviceID string) <-chan storage.Entry {
	ltr.mu.Lock()
	defer ltr.mu.Unlock()

	viewer := make(chan storage.Entry, viewerBufferCapacity)

	if ltr.Registry == nil {
		ltr.Registry = make(map[string][]chan storage.Entry)
	}

	ltr.Registry[serviceID] = append(ltr.Registry[serviceID], viewer)
	return viewer
}

func (ltr *liveTailRegistry) Unsubscribe(serviceID string, viewer <-chan storage.Entry) {

}

func (ltr *liveTailRegistry) Broadcast(entries []storage.Entry) {

}
