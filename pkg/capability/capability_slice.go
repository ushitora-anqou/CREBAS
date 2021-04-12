package capability

type CapabilitySlice []*Capability
type CapabilityRequestSlice []*CapabilityRequest

// Where returns a new CapabilitySlice whose elements return true for func
func (rcv CapabilitySlice) Where(fn func(*Capability) bool) (result CapabilitySlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Where returns a new CapabilityRequestSlice whose elements return true for func
func (rcv CapabilityRequestSlice) Where(fn func(*CapabilityRequest) bool) (result CapabilityRequestSlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
