package tkt

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

var defaultLogger = log.New(os.Stdout, ":", log.Ldate|log.Lshortfile)

type LogWriter struct {
	io.Writer
	FileName    string
	MaxSize     int
	MaxFiles    int
	inizialized bool
	fileNumber  int
	totalBytes  int
	file        *os.File
	mux         *sync.Mutex
	excludes    []string
}

func (o *LogWriter) Write(p []byte) (n int, err error) {
	if !o.inizialized {
		o.mux = &sync.Mutex{}
		o.initialize()
	}
	s1 := string(p)
	s2 := string(debug.Stack())
	for _, e := range o.excludes {
		if !strings.HasPrefix(s1, "error:") && strings.Contains(s2, e) {
			return
		}
	}
	o.mux.Lock()
	defer o.mux.Unlock()
	w, err := o.file.Write(p)
	if err != nil {
		return w, err
	}
	o.totalBytes += w
	if o.totalBytes >= o.MaxSize {
		o.file.Close()
		o.fileNumber++
		if o.fileNumber == o.MaxFiles {
			o.fileNumber = 1
		}
		o.createFile()
	}
	return w, nil
}

func (o *LogWriter) initialize() {
	o.mux.Lock()
	defer o.mux.Unlock()
	if o.inizialized {
		return
	}
	o.fileNumber = 1
	o.createFile()
	o.inizialized = true
}

func (o *LogWriter) createFile() {
	var err error
	name := fmt.Sprintf(o.FileName+".%d", o.fileNumber)

	o.file, err = os.Create(name)
	CheckErr(err)
	CheckErr(os.Chmod(name, 0644))

	o.totalBytes = 0
}

type NullWriter struct {
	io.Writer
}

func (o *NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type LoggersConfig struct {
	FileName     *string  `json:"fileName"`
	MaxSize      *int     `json:"maxSize"`
	MaxFiles     *int     `json:"maxFiles"`
	LogToConsole *bool    `json:"logToConsole"`
	Tags         []string `json:"tags"`
	Excludes     []string `json:"excludes"`
}

func (o *LoggersConfig) Validate() {
	if o.FileName == nil {
		panic("missing fileName")
	}
	if o.MaxSize == nil {
		panic("Missing maxSize")
	}
	if o.MaxFiles == nil {
		panic("Missing maxFile")
	}
	if o.LogToConsole == nil {
		panic("Missing logToConsole")
	}
	if o.Tags == nil {
		panic("Missing tags")
	}
	if o.Excludes == nil {
		panic("Missing excludes")
	}
}

type Loggers struct {
	config     *LoggersConfig
	output     io.Writer
	logWriter  *LogWriter
	loggerMap  map[string]*log.Logger
	nullLogger *log.Logger
}

func (o *Loggers) Config(config LoggersConfig) {
	w := &LogWriter{FileName: *config.FileName, MaxFiles: *config.MaxFiles, MaxSize: *config.MaxSize, excludes: config.Excludes}
	if *config.LogToConsole {
		o.output = io.MultiWriter(w, os.Stdout)
		log.SetOutput(o.output)
	} else {
		o.output = w
		log.SetOutput(w)
	}
	o.logWriter = w

	o.loggerMap = make(map[string]*log.Logger)
	for i := range config.Tags {
		prefix := config.Tags[i]
		o.loggerMap[prefix] = log.New(o.output, prefix+": ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	nullWriter := NullWriter{}
	o.nullLogger = log.New(&nullWriter, "", 0)
}

func (o *Loggers) Log(prefix string) *log.Logger {
	if o.loggerMap == nil {
		return defaultLogger
	} else if l, ok := o.loggerMap[prefix]; ok {
		return l
	} else {
		return o.nullLogger
	}
}

var loggers Loggers

func ConfigLoggers(fileName string, maxSize int, maxFiles int, console bool, tags ...string) *Loggers {
	loggers.Config(LoggersConfig{FileName: &fileName,
		MaxSize: &maxSize, MaxFiles: &maxFiles, LogToConsole: &console, Tags: tags})
	return &loggers
}

func InitLoggers(config LoggersConfig) *Loggers {
	loggers.Config(config)
	return &loggers
}

func Logger(tag string) *log.Logger {
	return loggers.Log(tag)
}

func Loggable(tag string) bool {
	_, ok := loggers.loggerMap[tag]
	return ok
}
