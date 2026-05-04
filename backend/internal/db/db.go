package db

// DB will be initialised in Task 2 (models + migrations).
type DB struct{}

func New(path string) (*DB, error) {
	return &DB{}, nil
}

func (d *DB) Close() error {
	return nil
}
