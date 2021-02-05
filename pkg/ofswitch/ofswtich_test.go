package ofswitch

import (
	"testing"

	"github.com/mattn/go-pipeline"
)

func TestCreate(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	_, err = pipeline.Output(
		[]string{"ip", "addr", "show"},
		[]string{"grep", ofs.Name},
	)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
}
