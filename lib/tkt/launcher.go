package tkt

import (
	"time"
)

type LauncherCallback func()

type callbackEntry struct {
	name     *string
	callback LauncherCallback
}

type Launcher struct {
	callbacks []callbackEntry
}

func (o *Launcher) Register(name string, callback LauncherCallback) {
	o.callbacks = append(o.callbacks, callbackEntry{name: &name, callback: callback})
}

func (o *Launcher) Launch() {
	go o.doLaunch()
}

func (o *Launcher) doLaunch() {
	defer func() {
		if r := recover(); r != nil {
			ProcessPanic(r)
		}
	}()
	for _, e := range o.callbacks {
		time.Sleep(2 * time.Second)
		Logger("info").Printf("Launching %s", *e.name)
		e.callback()
	}
}

func NewLauncher() *Launcher {
	return &Launcher{callbacks: make([]callbackEntry, 0)}
}
