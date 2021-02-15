package app

// AppSlice is a slice of type AppInterface
type AppSlice []AppInterface

// Where returns a new AppSlice whose elements return true for func
func (rcv AppSlice) Where(fn func(AppInterface) bool) (result AppSlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
