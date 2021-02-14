package app

// AppInterface is interface for application
type AppInterface interface {
	Start() error
	Stop() error
}
