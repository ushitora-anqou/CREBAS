package capability

type AppCertificateSlice []*AppCertificate

// Where returns a new UserSlice whose elements return true for func
func (rcv AppCertificateSlice) Where(fn func(*AppCertificate) bool) (result AppCertificateSlice) {
	for _, v := range rcv {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
