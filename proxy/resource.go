package proxy

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)
// resource is a single resource. It could be either local cache file or a remote url
// This object is created on each request.
type resource struct {
	URL            string
	Path           string
	Header         http.Header
	RemoteModified time.Time
}
// newResource creates new resource object with given url
func newResource(url string) (*resource, error) {
	if url == "" {
		return nil, errors.New("URL can't be empty")
	}
	path := strings.TrimPrefix(url, proxyURL)
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimPrefix(path, "/")
	return &resource{
		URL:  url,
		Path: dataDir + string(os.PathSeparator) + path,
	}, nil
}
// getRemoteLastModified sends HEAD request to resource URL and returns time in Last-Modified header.
func (rs *resource) getRemoteLastModified() (time.Time, int, error) {
	resp, err := http.Head(rs.URL)
	if err != nil {
		return time.Time{}, resp.StatusCode, err
	}
	defer resp.Body.Close()
	log.Infof("Status code %d", resp.StatusCode)
	rs.Header = resp.Header
	log.Infof("Got headers %+v", rs.Header)
	t, err := http.ParseTime(rs.Header.Get("Last-Modified"))
	rs.RemoteModified = t.UTC()
	return rs.RemoteModified, resp.StatusCode, err
}
// getRemote requests remote host for resource URL and returns response body reader and status code.
func (rs *resource) getRemote() (io.ReadCloser, int, error) {
	d.add(rs.URL)
	defer d.remove(rs.URL)
	log.WithField("url", rs.URL).Info("Getting remote file")
	resp, err := http.Get(rs.URL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, err
	}
	log.WithField("path", rs.Path).Info("Saving remote file")
	err = rs.saveLocal(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return rs.getLocal()
}
// getLocal looks into data dir for given resource and returns file reader
func (rs *resource) getLocal() (io.ReadCloser, int, error) {
	log.WithField("path", rs.Path).Info("Serving local file")
	file, err := os.Open(rs.Path)
	if err != nil {
		log.WithError(err).Errorf("Error opening file %s", rs.Path)
		return file, http.StatusInternalServerError, err
	}
	return file, http.StatusOK, nil
}
// saveLocal saves a reader (usually response body) in local file
// saved file's last modified time is same as remote file's last-modified header.
func (rs *resource) saveLocal(r io.ReadCloser) error {
	err := os.MkdirAll(filepath.Dir(rs.Path), 0711)
	if err != nil {
		return err
	}
	file, err := os.Create(rs.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(r)
	defer r.Close()
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err == nil {
		file.Close()
		log.Infof("Setting last modified to %s", rs.RemoteModified.String())
		err = os.Chtimes(rs.Path, rs.RemoteModified, rs.RemoteModified)
		if err != nil {
			log.WithError(err).Error("Couldn't change time")
		}
	}
	return nil
}
// getLocalLastModified runs stat on local file and returns last modified
func (rs *resource) getLocalLastModified() (time.Time, error) {
	statsinfo, err := os.Stat(rs.Path)
	if err != nil {
		return time.Time{}, err
	}
	return statsinfo.ModTime().UTC(), nil
}
