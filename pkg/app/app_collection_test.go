package app

import (
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	appCollection := NewAppCollection()
	app, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app.Stop()

	appCollection.Add(app)
	appTest := appCollection.GetByIndex(0)
	if app != appTest {
		t.Fatalf("Failed")
	}
}

func TestRemove(t *testing.T) {
	app, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app.Stop()
	app2, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app2.Stop()

	appCollection := NewAppCollection()
	appCollection.Add(app)
	appCollection.Add(app2)

	if count := appCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err = appCollection.Remove(app)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	linkTest := appCollection.GetByIndex(0)

	if linkTest != app2 {
		t.Fatalf("Failed")
	}

	if linkTest == app {
		t.Fatalf("Failed")
	}
}

func TestWhere(t *testing.T) {
	app, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app.Stop()

	app2, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app2.Stop()

	appCollection := NewAppCollection()
	appCollection.Add(app)
	appCollection.Add(app2)

	if count := appCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	slices := appCollection.Where(func(a AppInterface) bool {
		return a.ID() == app.ID()
	})

	if slices[0] == app2 {
		t.Fatalf("Failed")
	}

	if slices[0] != app {
		t.Fatalf("Failed")
	}
}

func TestClearNotRunningApp(t *testing.T) {
	app, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	app.cmd = []string{"/usr/bin/bash", "-c", "echo hello"}
	err = app.Start()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app.Stop()

	app2, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	app2.cmd = []string{"/usr/bin/bash", "-c", "while true; do sleep 1; done"}
	err = app2.Start()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app2.Stop()

	appCollection := NewAppCollection()
	appCollection.Add(app)
	appCollection.Add(app2)

	if count := appCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	time.Sleep(100 * time.Millisecond)
	err = appCollection.ClearNotRunningApp()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	linkTest := appCollection.GetByIndex(0)

	if !linkTest.IsRunning() {
		t.Fatalf("app is not running")
	}

	if linkTest != app2 {
		t.Fatalf("Failed expected:%v actual:%v", app2.ID(), linkTest.ID())
	}

	if linkTest.ID() == app.ID() {
		t.Fatalf("Failed")
	}
}
