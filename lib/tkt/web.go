package tkt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

type statsCounters struct {
	InCount         int64
	OutCount        int64
	TotalDuration   int64
	AverageDuration int64
}

func (o *statsCounters) Copy(other statsCounters) {
	o.InCount = other.InCount
	o.OutCount = other.OutCount
	o.TotalDuration = other.TotalDuration
	o.AverageDuration = other.AverageDuration
}

type pathStats struct {
	Path            *string
	IntervalCounter *statsCounters
	AccumCounters   *statsCounters
}

type webStats struct {
	mux   *sync.Mutex
	paths map[string]*pathStats
}

func (o *webStats) PushIn(path string) {
	o.mux.Lock()
	defer o.mux.Unlock()
	pathStats := o.resolvePathStats(path)
	pathStats.IntervalCounter.InCount++
	pathStats.AccumCounters.InCount++
}

func (o *webStats) PushOut(path string, duration int64) {
	o.mux.Lock()
	defer o.mux.Unlock()
	pathStats := o.resolvePathStats(path)
	pathStats.IntervalCounter.OutCount++
	pathStats.IntervalCounter.TotalDuration += duration
	pathStats.IntervalCounter.AverageDuration = int64(float64(pathStats.IntervalCounter.TotalDuration) / float64(pathStats.IntervalCounter.OutCount))
	pathStats.AccumCounters.OutCount++
	pathStats.AccumCounters.TotalDuration += duration
	pathStats.AccumCounters.AverageDuration = int64(float64(pathStats.AccumCounters.TotalDuration) / float64(pathStats.AccumCounters.OutCount))
}

func (o *webStats) resolvePathStats(path string) *pathStats {
	stats, ok := o.paths[path]
	if !ok {
		stats = &pathStats{
			Path: &path,
			IntervalCounter: &statsCounters{
				InCount:         0,
				OutCount:        0,
				AverageDuration: 0,
			},
			AccumCounters: &statsCounters{
				InCount:         0,
				OutCount:        0,
				AverageDuration: 0,
			},
		}
		o.paths[path] = stats
	}
	return stats
}

func (o *webStats) CheckPoint() []pathStats {
	o.mux.Lock()
	defer o.mux.Unlock()
	result := make([]pathStats, len(o.paths))
	i := 0
	for _, v := range o.paths {
		clone := pathStats{
			Path:            &*v.Path,
			IntervalCounter: &statsCounters{},
			AccumCounters:   &statsCounters{},
		}
		clone.IntervalCounter.Copy(*v.IntervalCounter)
		clone.AccumCounters.Copy(*v.AccumCounters)
		result[i] = clone
		v.IntervalCounter.InCount = 0
		v.IntervalCounter.OutCount = 0
		v.IntervalCounter.TotalDuration = 0
		v.IntervalCounter.AverageDuration = 0
		i++
	}
	return result
}

func (o *webStats) Start() {
	Logger("info").Println("Web services metering activated")
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			o.report()
		}
	}()
}

func (o *webStats) report() {
	defer func() {
		if r := recover(); r != nil {
			ProcessPanic(r)
		}
	}()
	snapshot := o.CheckPoint()
	sort.Slice(snapshot, func(i, j int) bool {
		return strings.Compare(*snapshot[i].Path, *snapshot[j].Path) > 0
	})
	buffer := bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("%-60s%15s%15s%15s%15s%15s%15s%15s%15s\r\n",
		"Path", "In", "Out", "Sum ms", "Avg ms", "Accum. In", "Out", "Sum ms", "Avg ms"))
	for _, v := range snapshot {
		buffer.WriteString(fmt.Sprintf("%-60s%15d%15d%15d%15d%15d%15d%15d%15d\r\n",
			*v.Path,
			v.IntervalCounter.InCount, v.IntervalCounter.OutCount, v.IntervalCounter.TotalDuration, v.IntervalCounter.AverageDuration,
			v.AccumCounters.InCount, v.AccumCounters.OutCount, v.AccumCounters.TotalDuration, v.AccumCounters.AverageDuration))
	}
	Logger("info").Printf("Web services stats:\r\n%s", buffer.String())
}

func newStats() *webStats {
	return &webStats{
		mux:   &sync.Mutex{},
		paths: make(map[string]*pathStats),
	}
}

var stats = newStats()

func InitWebStats() {
	stats.Start()
}

func ParseParamOrBody(r *http.Request, o interface{}) error {
	s := r.URL.Query().Get("body")
	if len(s) > 0 {
		return json.NewDecoder(strings.NewReader(s)).Decode(o)
	} else {
		return json.NewDecoder(r.Body).Decode(o)
	}
}

func InterceptCSP(delegate func(w http.ResponseWriter, r *http.Request), ancestors string) http.HandlerFunc {
	csp := fmt.Sprintf("frame-ancestors %s;", ancestors)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Security-Policy", csp)
		delegate(w, r)
	}
}

func InterceptCORS(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			header := r.Header.Get("Access-Control-Request-Headers")
			if len(header) > 0 {
				w.Header().Add("Access-Control-Allow-Headers", header)
			}
		} else {
			delegate(w, r)
		}
	}
}

func InterceptFatal(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return interceptStats(func(w http.ResponseWriter, r *http.Request) {
		defer catchFatal(w, r)
		delegate(w, r)
	})
}

func interceptStats(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		stats.PushIn(r.URL.Path)
		defer func() {
			t1 := time.Now()
			d := t1.Sub(t0).Milliseconds()
			stats.PushOut(r.URL.Path, d)
		}()
		delegate(w, r)
	}
}

func catchFatal(writer http.ResponseWriter, r *http.Request) {
	if e := recover(); e != nil {
		Logger("error").Printf("Error executing %s", r.URL.String())
		errorType := reflect.TypeOf(e)
		if errorType.Kind() == reflect.Struct || errorType.Kind() == reflect.Map || errorType.Kind() == reflect.Array {
			jsonBytes := Marshal(e)
			ProcessPanic(string(jsonBytes))
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(jsonBytes)
		} else {
			ProcessPanic(e)
			http.Error(writer, fmt.Sprint(e), http.StatusInternalServerError)
		}
	}
}

func JsonResponse(i interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "Application/json")
	JsonEncode(i, w)
}

func JsonErrorResponse(i interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusInternalServerError)
	JsonEncode(i, w)
}

type ReactResponseWriter struct {
	http.ResponseWriter
	notFound bool
}

func (o *ReactResponseWriter) WriteHeader(status int) {
	o.notFound = status == http.StatusNotFound
	if !o.notFound {
		o.ResponseWriter.WriteHeader(status)
	}
}

func (o *ReactResponseWriter) Write(b []byte) (int, error) {
	if o.notFound {
		return 0, nil
	} else {
		return o.ResponseWriter.Write(b)
	}
}

func InterceptReact(folder string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		interceptor := ReactResponseWriter{ResponseWriter: w, notFound: false}
		h.ServeHTTP(&interceptor, r)
		if interceptor.notFound {
			content, err := ioutil.ReadFile(folder + "/index.html")
			if err == nil {
				w.Header().Set("Content-Type", "text/html")
				_, err := w.Write(content)
				CheckErr(err)
			}
		}
	}
}

func LoadStaticContext(folder string, path string) {
	if FileExists(folder) {
		abs, _ := filepath.Abs(folder)
		println(abs)
		fs := http.FileServer(http.Dir(abs))
		sp := http.StripPrefix(path, fs)
		http.Handle(path, InterceptFatal(InterceptCORS(InterceptReact(folder, sp))))
	} else {
		log.Println("static content folder " + folder + " not found")
	}
}

func LoadStaticContextWithCSP(folder string, path string, ancestor string) {
	if FileExists(folder) {
		abs, _ := filepath.Abs(folder)
		println(abs)
		fs := http.FileServer(http.Dir(abs))
		sp := http.StripPrefix(path, fs)
		http.Handle(path, InterceptFatal(InterceptCORS(InterceptCSP(InterceptReact(folder, sp), ancestor))))
	} else {
		log.Println("static content folder " + folder + " not found")
	}
}

type ErrorResponse struct {
	ErrorMessage string      `json:"errorMessage"`
	Error        interface{} `json:"error"`
}
