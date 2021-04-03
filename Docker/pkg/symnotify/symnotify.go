// package symnotify provides a file system watcher that notifies events for symlink targets.
//
package symnotify

import (
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
        "strings"
	"github.com/fsnotify/fsnotify"
)

type Event = fsnotify.Event
type Op = fsnotify.Op

var debugOn bool = true

const (
	Create Op = fsnotify.Create
	Write     = fsnotify.Write
	Remove    = fsnotify.Remove
	Rename    = fsnotify.Rename
	Chmod     = fsnotify.Chmod
)

// Watcher is like fsnotify.Watcher but also notifies on changes to symlink targets
type Watcher struct {
	watcher *fsnotify.Watcher
	added   map[string]bool
}

func debug(f string, x ...interface{}) {
        if debugOn {
                log.Printf(f, x...)
        }
}

func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	return &Watcher{watcher: w, added: map[string]bool{}}, err
}

// Event returns the next event.
func (w *Watcher) Event() (e Event, err error) {
	return w.EventTimeout(time.Duration(math.MaxInt64))
}

// EventTimeout returns the next event or os.ErrDeadlineExceeded if timeout is exceeded.
func (w *Watcher) EventTimeout(timeout time.Duration) (e Event, err error) {
	var ok bool
	select {
	case e, ok = <-w.watcher.Events:
	case err, ok = <-w.watcher.Errors:
	case <-time.After(timeout):
		return Event{}, os.ErrDeadlineExceeded
	}
	switch {
	case !ok:
		return Event{}, io.EOF
	case e.Op == Create:
		debug("Create Event Detected for file e.Name %v",e.Name)
		if info, err := os.Lstat(e.Name); err == nil {
			if isSymlink(info) {
				_ = w.watcher.Add(e.Name)
			}
		}
	case e.Op == Remove:
		debug("Remove Event Detected for file e.Name %v",e.Name)
		w.watcher.Remove(e.Name)
	case e.Op == Chmod:
		debug("Chmod Event Detected for file e.Name %v",e.Name)
		if info, err := os.Lstat(e.Name); err == nil {
			if isSymlink(info) {
				// Symlink target may have changed.
				_ = w.watcher.Remove(e.Name)
				_ = w.watcher.Add(e.Name)
			}
		}
	}
	return e, err
}

// Add dir,dir/files* to the watcher
func (w *Watcher) Add(name string, containernames string) ([]string,error) {
	if err := w.watcher.Add(name); err != nil {
		return nil, err
	}
	w.added[name] = true // Explicitly added, don't auto-Remove

	// Scan directories for existing symlinks, we wont' get a Create for those.
        listofcontainernames := strings.Split(containernames," ")
        debug("list of container names %v",listofcontainernames)
        debug("name %v",name)
        debug("filenames %v",name+"/"+"*"+containernames+"*")
	matchedfilenames, err := filepath.Glob(name+"/"+"*"+containernames+"*") 
        debug("list of matched filenames %v",matchedfilenames)
	if (err != nil) {
        debug("no files found matching the container names %v",containernames)
        }
	//if files, err := ioutil.ReadDir(name); err == nil {
	//if files, err := ioutil.ReadDir(name); err == nil {
        for _, filename := range matchedfilenames {
		info,err :=  os.Lstat(filename) 
        	if (err == nil) {
		if isSymlink(info) {
		debug("Add file to watcher %v",filepath.Join(name, info.Name()))
				_ = w.watcher.Add(filepath.Join(name, info.Name()))
		}
		} 
	}
	return matchedfilenames, nil
}

// Remove name from watcher
func (w *Watcher) Remove(name string) error {
	delete(w.added, name)
	return w.watcher.Remove(name)
}

// Close watcher
func (w *Watcher) Close() error { return w.watcher.Close() }

func isSymlink(info os.FileInfo) bool {
	return (info.Mode() & os.ModeSymlink) == os.ModeSymlink
}
