package proxy

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	d        = &downloads{}
	proxyURL = ""
	dataDir  = ""
)
// Start will start a proxy server
// New server will start serving {url} on {host}:{port}/suffix. All cache will be saved in {dir}
// This function blocks unless an error occurred.
func Start(url, suffix, dir, host string, port int) error {
	if url == "" || dir == "" {
		return errors.New("proxyURL and dataDir are required parameters and can't be blank.")
	}
	proxyURL = url
	dataDir = dir
	r := mux.NewRouter()
	r.HandleFunc(suffix+"/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Got request from %s", r.RemoteAddr)
		vars := mux.Vars(r)
		path := vars["path"]
		queryParams := r.URL.RawQuery
		if queryParams != "" {
			path = path + "?" + queryParams
		}
		log.Infof("Fetching %s", url+path)
		reader, status, h, err := get( url + path)
		if err != nil || status != http.StatusOK {
			log.WithError(err).Error("Error occurred")
			http.Error(w, http.StatusText(status), status)
			return
		}
		header := w.Header()
		for key, val := range h {
			header[key] = val
		}
		b, err := ioutil.ReadAll(reader)
		defer reader.Close()
		if err != nil {
			log.WithError(err).Error("Error occurred")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})
	return http.ListenAndServe(
		fmt.Sprintf("%s:%d", host, port),
		handlers.CombinedLoggingHandler(os.Stdout, r))
}

// get tries to serve a requested resource url.
// A HEAD request is sent to remote host to find out Last-Modified header
// If we have local file and its last modified is same or greater than remote, local file is served
// If remote has newer file, we save it locally and then serve, so future hits will be served from cache.
// When a remote file is being downloaded, all parallel requests for it are blocked until file is locally available, then all are served by local cache.
func get(rsURL string) (io.ReadCloser, int, http.Header, error) {
	c, err := newResource(rsURL)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, err
	}
	remoteMod, status, err := c.getRemoteLastModified()
	if status != http.StatusOK {
		return nil, status, c.Header, err
	}
	// wait so any pending downloads may finish
	d.waitFor(rsURL)
	localMod, err := c.getLocalLastModified()
	if err != nil {
		log.WithError(err).Warn("Couldn't find local last modified.")
	}
	log.Infof("Local Modified: %s, Remote Modified: %s", localMod.String(), remoteMod.String())
	if localMod.UnixNano() < remoteMod.UnixNano() {
		r, status, err := c.getRemote()
		if err != nil {
			log.WithError(err).Error("Couldn't download remote file")
		}
		return r, status, c.Header, err
	} else {
		log.WithField("rsURL", c.URL).Info("Cache hit")
		r, status, err := c.getLocal()
		return r, status, c.Header, err
	}
}
