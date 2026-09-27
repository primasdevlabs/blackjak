package storage

// Storage defines an interface for data persistence.
type Storage interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}
