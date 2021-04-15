package app

// AppSlice is a slice of type AppInterface
type DeviceSlice []*Device

// Where returns a new AppSlice whose elements return true for func
func (rcv DeviceSlice) Where(fn func(*Device) bool) (result DeviceSlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
