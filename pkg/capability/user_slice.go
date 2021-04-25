package capability

type UserGrantPolicySlice []*UserGrantPolicy

// Where returns a new UserSlice whose elements return true for func
func (rcv UserGrantPolicySlice) Where(fn func(*UserGrantPolicy) bool) (result UserGrantPolicySlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
