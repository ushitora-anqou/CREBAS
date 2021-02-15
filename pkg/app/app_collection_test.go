package app

import (
	"testing"
)

func TestAdd(t *testing.T) {
	appCollection := NewAppCollection()
	app := NewLinuxProcess()

	appCollection.Add(app)
	appTest := appCollection.GetByIndex(0)
	if app != appTest {
		t.Fatalf("Failed")
	}
}

func TestRemove(t *testing.T) {
	app := NewLinuxProcess()
	app2 := NewLinuxProcess()

	appCollection := NewAppCollection()
	appCollection.Add(app)
	appCollection.Add(app2)

	if count := appCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err := appCollection.Remove(app)
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
	app := NewLinuxProcess()
	err := app.Create()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app.Delete()

	app2 := NewLinuxProcess()
	err = app2.Create()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer app2.Delete()

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
