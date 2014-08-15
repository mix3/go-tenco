package storage

type Storager interface {
	Map() (map[string]string, error)
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
	DeleteAll() error
	Close()
}
