package main

import (
	"log"
	"os"
	"github.com/pmoogi-redhat/go-log-file-watcher/pkg/symnotify"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
)


var  debugOn bool = true
var  containernames string = ""

type FileWatcher struct {
        watcher *symnotify.Watcher
        metrics *prometheus.CounterVec
        sizes map[string]float64
}


func (w *FileWatcher) Update(path string) error {
        counter, err := w.metrics.GetMetricWithLabelValues(path)
        if err != nil {
                return err
        }
        stat, err := os.Stat(path)
        if err != nil {
                return err
        }
        if stat.IsDir() {
                return nil // Ignore directories
        }
        lastSize, size := float64(w.sizes[path]), float64(stat.Size())
        w.sizes[path] = size
        var add float64
        if size > lastSize {
                // File has grown, add the difference to the counter.
                add = size - lastSize
        } else if size < lastSize {
                // File truncated, starting over. Add the size.
                add = size
        }
        debug("%v: (%v->%v) +%v", path, lastSize, size, add)
        counter.Add(add)
        return nil
}

func (w FileWatcher) Watch() {

        for {
                e, err := w.watcher.Event()
                fatal(err)
                if (strings.Contains(e.Name,".log")) {
                if (strings.Contains(e.Name,containernames)) {
                debug("Event notified for e.Name %v call w.Update", e.Name)
                w.Update(e.Name)
                }}
        }
}


func main() {
        var dir string
        var listeningport string

	//directory to be watched out where symlinks to all logs files are present e.g. /var/log/containers/
	//debug option true or false
	//listening port where this go-app push prometheus registered metrics for further collected or reading by end prometheus server
	flag.StringVar(&dir, "logfilespathname", "/var/log/containers/", "Give the dirname where logfiles are going to be located, default /var/log/containers/")
	flag.StringVar(&containernames,"containernames","log-stress","Given container names e.g. xxx yyy zzz only their log files are followed default is low-stress")
	flag.BoolVar(&debugOn, "debug", false, "Give debug option false or true, default set to true")
	flag.StringVar(&listeningport, "listeningport", ":2112", "Give the listening port where metrics can be exposed to and listened by a running prometheus server, default is :2112")
	flag.Parse()

	debug("logfilespathname= %v",dir)
	debug("containernames= %v",containernames)
	debug("debug option= %v",debugOn)
	debug("listening port address= %v",listeningport)


	//Get new watcher
        w := &FileWatcher{
                metrics: prometheus.NewCounterVec(prometheus.CounterOpts{
                        Name: "fluentd_input_status_total_bytes_logged",
                        Help: "total bytes logged to disk (log file) ",
                }, []string{"path"}),
                sizes: make(map[string]float64),
        }
	defer prometheus.Register(w.metrics)

	w.watcher, err := symnotify.NewFileWatcher(dir)
	if err != nil {
		debug("NewFileWatcher error")
		log.Fatal(err)
	}
	//Add dir to watcher
	w.watcher.Add(dir)
	w.watcher.added[dir] = true

	go w.Watch()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listeningport, nil)
}