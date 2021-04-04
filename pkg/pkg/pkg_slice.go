package pkg

// AppSlice is a slice of type AppInterface
type PkgSlice []*PackageInfo

// Where returns a new AppSlice whose elements return true for func
func (rcv PkgSlice) Where(fn func(*PackageInfo) bool) (result PkgSlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
