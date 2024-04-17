package sqlite

import (
	"context"
	"fmt"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openfga/openfga/pkg/storage"
)

var _ storage.OpenFGADatastore = (*SQLite)(nil)
var _ storage.OpenFGADatastoreDriver = (*SQLiteDriver)(nil)

type SQLiteDriver struct{}

// Open implements storage.OpenFGADatastoreDriver.
func (s *SQLiteDriver) Open(uri string) (storage.OpenFGADatastore, error) {
	return &SQLite{}, nil
}

type SQLite struct{}

// New constructs and returns an implementation of a
// SQLite OpenFGADatastore.
func New() *SQLite {
	return &SQLite{}
}

// Close implements storage.OpenFGADatastore.
func (s *SQLite) Close() {
}

// CreateStore implements storage.OpenFGADatastore.
func (s *SQLite) CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error) {
	return nil, fmt.Errorf("not implemented")
}

// DeleteStore implements storage.OpenFGADatastore.
func (s *SQLite) DeleteStore(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

// FindLatestAuthorizationModel implements storage.OpenFGADatastore.
func (s *SQLite) FindLatestAuthorizationModel(ctx context.Context, store string) (*openfgav1.AuthorizationModel, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetStore implements storage.OpenFGADatastore.
func (s *SQLite) GetStore(ctx context.Context, id string) (*openfgav1.Store, error) {
	return nil, fmt.Errorf("not implemented")
}

// IsReady implements storage.OpenFGADatastore.
func (s *SQLite) IsReady(ctx context.Context) (storage.ReadinessStatus, error) {
	return storage.ReadinessStatus{}, fmt.Errorf("not implemented")
}

// ListStores implements storage.OpenFGADatastore.
func (s *SQLite) ListStores(ctx context.Context, paginationOptions storage.PaginationOptions) ([]*openfgav1.Store, []byte, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// MaxTuplesPerWrite implements storage.OpenFGADatastore.
func (s *SQLite) MaxTuplesPerWrite() int {
	return 0
}

// MaxTypesPerAuthorizationModel implements storage.OpenFGADatastore.
func (s *SQLite) MaxTypesPerAuthorizationModel() int {
	return 0
}

// Read implements storage.OpenFGADatastore.
func (s *SQLite) Read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (storage.Iterator[*openfgav1.Tuple], error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadAssertions implements storage.OpenFGADatastore.
func (s *SQLite) ReadAssertions(ctx context.Context, store string, modelID string) ([]*openfgav1.Assertion, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadAuthorizationModel implements storage.OpenFGADatastore.
func (s *SQLite) ReadAuthorizationModel(ctx context.Context, store string, id string) (*openfgav1.AuthorizationModel, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadAuthorizationModels implements storage.OpenFGADatastore.
func (s *SQLite) ReadAuthorizationModels(ctx context.Context, store string, options storage.PaginationOptions) ([]*openfgav1.AuthorizationModel, []byte, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ReadChanges implements storage.OpenFGADatastore.
func (s *SQLite) ReadChanges(ctx context.Context, store string, objectType string, paginationOptions storage.PaginationOptions, horizonOffset time.Duration) ([]*openfgav1.TupleChange, []byte, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ReadPage implements storage.OpenFGADatastore.
func (s *SQLite) ReadPage(ctx context.Context, store string, tupleKey *openfgav1.TupleKey, paginationOptions storage.PaginationOptions) ([]*openfgav1.Tuple, []byte, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// ReadStartingWithUser implements storage.OpenFGADatastore.
func (s *SQLite) ReadStartingWithUser(ctx context.Context, store string, filter storage.ReadStartingWithUserFilter) (storage.Iterator[*openfgav1.Tuple], error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadUserTuple implements storage.OpenFGADatastore.
func (s *SQLite) ReadUserTuple(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (*openfgav1.Tuple, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadUsersetTuples implements storage.OpenFGADatastore.
func (s *SQLite) ReadUsersetTuples(ctx context.Context, store string, filter storage.ReadUsersetTuplesFilter) (storage.Iterator[*openfgav1.Tuple], error) {
	return nil, fmt.Errorf("not implemented")
}

// Write implements storage.OpenFGADatastore.
func (s *SQLite) Write(ctx context.Context, store string, d []*openfgav1.TupleKeyWithoutCondition, w []*openfgav1.TupleKey) error {
	return fmt.Errorf("not implemented")
}

// WriteAssertions implements storage.OpenFGADatastore.
func (s *SQLite) WriteAssertions(ctx context.Context, store string, modelID string, assertions []*openfgav1.Assertion) error {
	return fmt.Errorf("not implemented")
}

// WriteAuthorizationModel implements storage.OpenFGADatastore.
func (s *SQLite) WriteAuthorizationModel(ctx context.Context, store string, model *openfgav1.AuthorizationModel) error {
	return fmt.Errorf("not implemented")
}
