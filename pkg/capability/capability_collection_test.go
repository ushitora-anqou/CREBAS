package capability

import "testing"

func TestCapabilityCollectionRemove(t *testing.T) {
	cap := NewCreateSkeltonCapability()
	cap2 := NewCreateSkeltonCapability()

	capCollection := NewCapabilityCollection()
	capCollection.Add(cap)
	capCollection.Add(cap2)

	if count := capCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err := capCollection.Remove(cap)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	linkTest := capCollection.GetByIndex(0)

	if linkTest != cap2 {
		t.Fatalf("Failed")
	}

}

func TestCapabilityCollectionWhere(t *testing.T) {
	cap := NewCreateSkeltonCapability()
	cap2 := NewCreateSkeltonCapability()

	capCollection := NewCapabilityCollection()
	capCollection.Add(cap)
	capCollection.Add(cap2)

	if count := capCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	slices := capCollection.Where(func(a *Capability) bool {
		return a.CapabilityID == cap.CapabilityID
	})

	if slices[0] == cap2 {
		t.Fatalf("Failed")
	}

	if slices[0] != cap {
		t.Fatalf("Failed")
	}
}

func TestCapabilityRequestCollectionRemove(t *testing.T) {
	capReq := NewCreateSkeltonCapabilityRequest()
	capReq2 := NewCreateSkeltonCapabilityRequest()

	capCollection := NewCapabilityRequestCollection()
	capCollection.Add(capReq)
	capCollection.Add(capReq2)

	if count := capCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err := capCollection.Remove(capReq)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	linkTest := capCollection.GetByIndex(0)

	if linkTest != capReq2 {
		t.Fatalf("Failed")
	}

}

func TestCapabilityRequestCollectionWhere(t *testing.T) {
	capReq := NewCreateSkeltonCapabilityRequest()
	capReq2 := NewCreateSkeltonCapabilityRequest()

	capCollection := NewCapabilityRequestCollection()
	capCollection.Add(capReq)
	capCollection.Add(capReq2)

	if count := capCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	slices := capCollection.Where(func(a *CapabilityRequest) bool {
		return a.RequestID == capReq.RequestID
	})

	if slices[0] == capReq2 {
		t.Fatalf("Failed")
	}

	if slices[0] != capReq {
		t.Fatalf("Failed")
	}
}
