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

func TestCreate(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
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
