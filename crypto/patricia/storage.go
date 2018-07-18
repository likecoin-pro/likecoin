package patricia

type Storage interface {
	Get(key []byte) ([]byte, error)
	Put(key, data []byte) error
	//Delete(key []byte) error
}

type Reader interface {
	Get(key []byte) ([]byte, error)
}

type MemoryStorage struct {
	reader Reader
	data   map[string][]byte
}

func NewMemoryStorage(reader Reader) *MemoryStorage {
	return &MemoryStorage{
		reader: reader,
		data:   map[string][]byte{},
	}
}

func (s *MemoryStorage) Get(key []byte) (v []byte, err error) {
	v, ok := s.data[string(key)]
	if !ok && s.reader != nil {
		if v, err = s.reader.Get(key); err == nil && v != nil {
			s.data[string(key)] = v
		}
	}
	return
}

func (s *MemoryStorage) Put(key, val []byte) error {
	s.data[string(key)] = val
	return nil
}

type subStorage struct {
	s      Storage
	prefix []byte
}

func NewSubStorage(s Storage, keyPrefix []byte) Storage {
	return &subStorage{s, keyPrefix}
}

func (s *subStorage) Get(key []byte) (v []byte, err error) {
	return s.s.Get(append(s.prefix, key...))
}

func (s *subStorage) Put(key, val []byte) error {
	return s.s.Put(append(s.prefix, key...), val)
}
