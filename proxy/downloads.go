package proxy

import (
	"sync"
	"time"
)
// Downloads is bucket of in progress remote requests for resources
// This is necessary to stop making parallel requests for same resource when its not available locally.
type downloads struct {
	list map[string]int
	sync.Mutex
}
// has checks if download bucket has given url in progress
func (d *downloads) has(rsURL string) bool {
	if _, ok := d.list[rsURL]; ok {
		return true
	}
	return false
}
// remove removes given url from active downloads. This indicates that download is finished and file is in local cache.
func (d *downloads) remove(rsURL string) {
	d.Lock()
	defer d.Unlock()
	delete(d.list, rsURL)
}
// add indicates file has started downloading and all requests for this file should block until download is no more in progress.
func (d *downloads) add(rsURL string) {
	d.Lock()
	defer d.Unlock()
	if d.list == nil {
		d.list = make(map[string]int, 1)
	}
	d.list[rsURL] = 1
}
// waitFor blocks until rsURL is no more in downloads bucket. Each check is delayed by one second.
func (d *downloads) waitFor(rsURL string) {
	if d.has(rsURL) {
		time.Sleep(time.Second)
		d.waitFor(rsURL)
	}
	return
}
