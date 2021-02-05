package ofswitch

import (
	"testing"

	"github.com/digitalocean/go-openvswitch/ovs"
)

func ovsBridgeExists(bridgeName string) (bool, error) {
	client := ovs.New()
	bridges, err := client.VSwitch.ListBridges()
	if err != nil {
		return false, err
	}

	for _, bridge := range bridges {
		if bridge == bridgeName {
			return true, nil
		}
	}

	return false, nil
}

func getOvsController(bridgeName string) (string, error) {
	client := ovs.New()
	return client.VSwitch.GetController(bridgeName)
}

func TestCreate(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	linkName := (*ofs.link).Attrs().Name
	if linkName != ofs.Name {
		t.Fatalf("failed test expected %v actual %v", ofs.Name, linkName)
	}

	exist, err := ovsBridgeExists(ofs.Name)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if !exist {
		t.Fatalf("failed test ovs bridge %v does not exist", ofs.Name)
	}
}

func TestCreateAndRemove(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.Remove()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	exist, err := ovsBridgeExists(ofs.Name)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if exist {
		t.Fatalf("failed test ovs bridge %v remained", ofs.Name)
	}
}

func TestSetController(t *testing.T) {
	ovsName := "ovs-test-set"
	ofs := NewOFSwitch(ovsName)
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetController("localhost:6655")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	controllerURL, err := getOvsController(ovsName)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if controllerURL != "localhost:6655" {
		t.Fatalf("failed to test expected localhost:6655 actual %v", controllerURL)
	}

	err = ofs.Remove()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
}
