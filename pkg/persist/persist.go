package persist

type Persistent interface {
	Save(key string, data interface{}) error
	Load(key string, data interface{}) (bool, error)
	Delete(key string) error
}
