# RFC: SQL Storage Provider Extensibility

## Status
Draft

## Summary
This RFC describes the extensibility approach for implementing custom SQL storage providers in OpenFGA. The approach uses a two-tier driver pattern combined with OpenFGA's plugin system to enable seamless integration of SQL databases as storage backends.

## Motivation
OpenFGA provides a flexible storage abstraction layer that allows custom implementations beyond the built-in PostgreSQL and MySQL support. Organizations may need to:
- Use alternative SQL databases (SQLite, CockroachDB, etc.)
- Implement custom storage optimizations
- Add database-specific features (partitioning, sharding, etc.)
- Support legacy or proprietary database systems

## Architecture Overview

### Two-Tier Pattern

SQL storage providers follow a two-tier architecture:

1. **Driver Tier**: Factory for creating datastore instances
   - Implements `storage.OpenFGADatastoreDriver`
   - Handles URI-based configuration
   - Manages connection initialization

2. **Datastore Tier**: Actual storage implementation
   - Implements `storage.OpenFGADatastore` (23 methods)
   - Executes database operations
   - Manages lifecycle (connections, cleanup)

### Plugin Registration

Storage providers register with OpenFGA through the plugin system:

```go
func InitPlugin(pm *plugin.PluginManager) error {
    driver := &MyDriver{}
    return pm.RegisterOpenFGADatastore("mydb", driver)
}
```

## Implementation Guide

### Step 1: Define Driver Structure

Create a driver that implements `storage.OpenFGADatastoreDriver`:

```go
package mydb

import (
    "database/sql"
    "github.com/openfga/openfga/pkg/storage"
)

type MyDriver struct{}

// Open creates a new datastore instance from a connection URI
func (d *MyDriver) Open(uri string) (storage.OpenFGADatastore, error) {
    db, err := sql.Open("mydb", uri)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    return &MyDatastore{
        db: db,
    }, nil
}
```

**Key Points**:
- The `uri` parameter contains database-specific connection info
- Parse and validate the URI in the `Open` method
- Initialize connection pool and database resources
- Return a configured datastore instance

### Step 2: Implement Datastore Interface

Create the datastore struct and implement all 23 required methods:

```go
type MyDatastore struct {
    db *sql.DB
}

// Ensure interface compliance at compile time
var _ storage.OpenFGADatastore = (*MyDatastore)(nil)
```

### Required Method Categories

#### A. Store Management (4 methods)

```go
func (ds *MyDatastore) CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error)
func (ds *MyDatastore) DeleteStore(ctx context.Context, id string) error
func (ds *MyDatastore) GetStore(ctx context.Context, id string) (*openfgav1.Store, error)
func (ds *MyDatastore) ListStores(ctx context.Context, opts storage.PaginationOptions) ([]*openfgav1.Store, []byte, error)
```

**Implementation Notes**:
- Store `id`, `name`, `created_at`, `updated_at`, `deleted_at` fields
- Use UUID for store IDs
- Support soft deletion (deleted_at)
- Implement continuation token pagination in `ListStores`

#### B. Authorization Models (4 methods)

```go
func (ds *MyDatastore) WriteAuthorizationModel(ctx context.Context, store string, model *openfgav1.AuthorizationModel) error
func (ds *MyDatastore) ReadAuthorizationModel(ctx context.Context, store string, id string) (*openfgav1.AuthorizationModel, error)
func (ds *MyDatastore) ReadAuthorizationModels(ctx context.Context, store string, opts storage.PaginationOptions) ([]*openfgav1.AuthorizationModel, []byte, error)
func (ds *MyDatastore) FindLatestAuthorizationModel(ctx context.Context, store string) (*openfgav1.AuthorizationModel, error)
```

**Implementation Notes**:
- Serialize `TypeDefinitions` as JSON or protocol buffer bytes
- Store model version/ID for retrieval
- Index by store ID and model ID
- `FindLatestAuthorizationModel` returns most recent by creation time

#### C. Assertions (2 methods)

```go
func (ds *MyDatastore) WriteAssertions(ctx context.Context, store, modelID string, assertions []*openfgav1.Assertion) error
func (ds *MyDatastore) ReadAssertions(ctx context.Context, store, modelID string) ([]*openfgav1.Assertion, error)
```

**Implementation Notes**:
- Assertions validate authorization models
- Store per model (store + modelID key)
- Replace all assertions on write (not append)

#### D. Tuple Operations (6 methods)

The core of OpenFGA - relationship tuples (subject-relation-object):

```go
func (ds *MyDatastore) Write(ctx context.Context, store string,
    deletes []*openfgav1.TupleKeyWithoutCondition,
    writes []*openfgav1.TupleKey) error

func (ds *MyDatastore) Read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (storage.Iterator[*openfgav1.Tuple], error)

func (ds *MyDatastore) ReadPage(ctx context.Context, store string, tupleKey *openfgav1.TupleKey, opts storage.PaginationOptions) ([]*openfgav1.Tuple, []byte, error)

func (ds *MyDatastore) ReadUserTuple(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (*openfgav1.Tuple, error)

func (ds *MyDatastore) ReadStartingWithUser(ctx context.Context, store string, filter storage.ReadStartingWithUserFilter) (storage.Iterator[*openfgav1.Tuple], error)

func (ds *MyDatastore) ReadUsersetTuples(ctx context.Context, store string, filter storage.ReadUsersetTuplesFilter) (storage.Iterator[*openfgav1.Tuple], error)
```

**Implementation Notes**:
- Tuples have: `object`, `relation`, `user`, optional `condition`
- The `Write` method is transactional (deletes + writes atomic)
- Implement efficient indexes for query patterns:
  - By object (object, relation)
  - By user (for reverse lookups)
  - By object type
- `Read` returns iterator for streaming large result sets
- `ReadPage` uses continuation tokens for pagination
- `ReadUserTuple` finds exact match
- `ReadStartingWithUser` supports user-based queries
- `ReadUsersetTuples` finds tuples where user is a userset (e.g., "group#member")

#### E. Change Tracking (1 method)

```go
func (ds *MyDatastore) ReadChanges(ctx context.Context, store, objectType string,
    opts storage.PaginationOptions, horizonOffset time.Duration) ([]*openfgav1.TupleChange, []byte, error)
```

**Implementation Notes**:
- Track all tuple changes (INSERT, DELETE operations)
- Store: operation type, tuple data, timestamp
- Filter by object type
- `horizonOffset` limits how far back changes are retained
- Support pagination with continuation tokens

#### F. Configuration & Status (4 methods)

```go
func (ds *MyDatastore) MaxTuplesPerWrite() int
func (ds *MyDatastore) MaxTypesPerAuthorizationModel() int
func (ds *MyDatastore) IsReady(ctx context.Context) (storage.ReadinessStatus, error)
func (ds *MyDatastore) Close()
```

**Implementation Notes**:
- `MaxTuplesPerWrite`: Return database-specific limit (e.g., 100)
- `MaxTypesPerAuthorizationModel`: Return limit (e.g., 100)
- `IsReady`: Perform health check (ping database)
- `Close`: Clean up connections and resources

### Step 3: Schema Design

Design SQL schema optimized for OpenFGA query patterns:

**Stores Table**:
```sql
CREATE TABLE store (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);
```

**Authorization Models Table**:
```sql
CREATE TABLE authorization_model (
    store_id VARCHAR(255) NOT NULL,
    model_id VARCHAR(255) NOT NULL,
    schema_version VARCHAR(20) NOT NULL,
    type_definitions BLOB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (store_id, model_id),
    FOREIGN KEY (store_id) REFERENCES store(id)
);

CREATE INDEX idx_auth_model_created ON authorization_model(store_id, created_at DESC);
```

**Tuples Table** (Critical for Performance):
```sql
CREATE TABLE tuple (
    store_id VARCHAR(255) NOT NULL,
    object_type VARCHAR(255) NOT NULL,
    object_id VARCHAR(255) NOT NULL,
    relation VARCHAR(50) NOT NULL,
    user_object_type VARCHAR(255) NOT NULL,
    user_object_id VARCHAR(255) NOT NULL,
    user_relation VARCHAR(50),
    condition_name VARCHAR(255),
    condition_context BLOB,
    ulid VARCHAR(26) NOT NULL,
    inserted_at TIMESTAMP NOT NULL,
    PRIMARY KEY (store_id, object_type, object_id, relation, user_object_type, user_object_id, user_relation),

    -- Critical indexes for query performance
    INDEX idx_tuple_object (store_id, object_type, object_id, relation),
    INDEX idx_tuple_user (store_id, user_object_type, user_object_id),
    INDEX idx_tuple_object_type (store_id, object_type)
);
```

**Changelog Table**:
```sql
CREATE TABLE changelog (
    store_id VARCHAR(255) NOT NULL,
    object_type VARCHAR(255) NOT NULL,
    object_id VARCHAR(255) NOT NULL,
    relation VARCHAR(50) NOT NULL,
    user_object_type VARCHAR(255) NOT NULL,
    user_object_id VARCHAR(255) NOT NULL,
    user_relation VARCHAR(50),
    operation VARCHAR(10) NOT NULL, -- 'INSERT' or 'DELETE'
    timestamp TIMESTAMP NOT NULL,
    ulid VARCHAR(26) NOT NULL,
    PRIMARY KEY (store_id, ulid),
    INDEX idx_changelog_type_time (store_id, object_type, timestamp)
);
```

**Assertions Table**:
```sql
CREATE TABLE assertion (
    store_id VARCHAR(255) NOT NULL,
    model_id VARCHAR(255) NOT NULL,
    assertions BLOB NOT NULL,
    PRIMARY KEY (store_id, model_id),
    FOREIGN KEY (store_id, model_id) REFERENCES authorization_model(store_id, model_id)
);
```

### Step 4: Create Plugin Entry Point

Create plugin registration in `plugins/storage/mydb/main.go`:

```go
package main

import (
    "github.com/openfga/openfga/pkg/plugin"
    "github.com/yourusername/openfga-contrib/storage/mydb"
)

// InitPlugin is called by OpenFGA to initialize the plugin
func InitPlugin(pm *plugin.PluginManager) error {
    driver := &mydb.MyDriver{}
    return pm.RegisterOpenFGADatastore("mydb", driver)
}
```

### Step 5: Implement Conformance Tests

Validate implementation using OpenFGA's conformance test suite:

```go
package mydb

import (
    "testing"
    "github.com/openfga/openfga/pkg/storage/test"
)

func TestOpenFGA_StorageConformance(t *testing.T) {
    driver := &MyDriver{}
    ds, err := driver.Open("mydb://localhost/testdb")
    if err != nil {
        t.Fatalf("failed to create datastore: %v", err)
    }
    defer ds.Close()

    // Run complete conformance test suite
    test.RunAllTests(t, ds)
}
```

**The conformance suite validates**:
- Store CRUD operations
- Authorization model lifecycle
- Tuple write/read operations
- Query filtering and pagination
- Change tracking
- Concurrent operations
- Edge cases and error handling

## Performance Considerations

### Query Optimization

1. **Index Strategy**:
   - Primary index: (store_id, object_type, object_id, relation, user)
   - Object lookups: (store_id, object_type, object_id, relation)
   - User lookups: (store_id, user_object_type, user_object_id)
   - Type filtering: (store_id, object_type)

2. **Batch Operations**:
   - Use prepared statements for bulk inserts
   - Batch delete operations in single transaction
   - Consider using COPY/BULK INSERT for large datasets

3. **Connection Pooling**:
   - Configure appropriate pool size for workload
   - Set connection timeouts
   - Monitor connection utilization

4. **Pagination**:
   - Implement efficient continuation tokens (ULID-based)
   - Avoid OFFSET-based pagination for large datasets
   - Use indexed columns in pagination ORDER BY

### Scalability Patterns

1. **Read Replicas**: Direct read queries to replicas
2. **Partitioning**: Partition tuples table by store_id or object_type
3. **Caching**: Consider tuple result caching for hot paths
4. **Write Optimization**: Use write-ahead logging appropriately

## Error Handling

Implement consistent error handling:

```go
import (
    "errors"
    "github.com/openfga/openfga/pkg/storage"
)

// Map database errors to OpenFGA error types
if errors.Is(err, sql.ErrNoRows) {
    return storage.ErrNotFound
}
if isDuplicateKeyError(err) {
    return storage.ErrCollision
}
```

**Standard OpenFGA Errors**:
- `storage.ErrNotFound`: Resource doesn't exist
- `storage.ErrCollision`: Duplicate key/constraint violation
- `storage.ErrInvalidContinuationToken`: Bad pagination token
- `storage.ErrMismatchObjectType`: Type validation failure
- `storage.ErrTransactionalWriteFailed`: Transaction rollback

## Testing Strategy

1. **Unit Tests**: Test individual methods with mock data
2. **Conformance Tests**: Use OpenFGA's test suite (required)
3. **Integration Tests**: Test with real database instance
4. **Performance Tests**: Benchmark query patterns
5. **Concurrent Tests**: Validate thread safety and transactions

## Example: SQLite Implementation

See reference implementation at `storage/sqlite/sqlite.go`:

```go
package sqlite

import (
    "context"
    "github.com/openfga/openfga/pkg/storage"
)

type SQLiteDriver struct{}

func (s *SQLiteDriver) Open(uri string) (storage.OpenFGADatastore, error) {
    // Initialize SQLite connection
    return &SQLite{}, nil
}

type SQLite struct {
    // db *sql.DB
}

var _ storage.OpenFGADatastore = (*SQLite)(nil)

// Implement all 23 required methods...
```

## Migration and Versioning

1. **Schema Versioning**: Track schema version in metadata table
2. **Migration Scripts**: Provide SQL migration scripts for schema changes
3. **Backward Compatibility**: Maintain compatibility during upgrades
4. **Data Migration**: Plan for large dataset migrations

## Configuration

Storage providers receive configuration via URI:

```
mydb://host:port/database?param1=value1&param2=value2
```

Parse connection parameters in the `Open` method:
- Host and port
- Database name
- Authentication credentials
- Connection pool settings
- Database-specific options

## Deployment

1. **Build**: Compile as Go plugin or static binary
2. **Configuration**: Configure OpenFGA with storage URI
3. **Migration**: Run schema migration scripts
4. **Validation**: Verify with conformance tests
5. **Monitoring**: Track query performance and errors

## Security Considerations

1. **SQL Injection**: Use parameterized queries exclusively
2. **Credentials**: Store credentials securely (environment variables, secrets manager)
3. **Encryption**: Support TLS/SSL for database connections
4. **Audit**: Log security-relevant operations
5. **Access Control**: Use database-level access controls

## Future Enhancements

Potential future improvements:
- Connection pooling strategies
- Query result caching layer
- Distributed transaction support
- Multi-region replication patterns
- Advanced partitioning strategies

## References

- OpenFGA Storage Interface: `github.com/openfga/openfga/pkg/storage`
- Plugin System: `github.com/openfga/openfga/pkg/plugin`
- Conformance Tests: `github.com/openfga/openfga/pkg/storage/test`
- Reference Implementation: `storage/sqlite/sqlite.go`

## Conclusion

The SQL Storage Provider extensibility approach provides a clear path for implementing custom database backends while ensuring compatibility through:
1. Well-defined interfaces (Driver + Datastore)
2. Plugin-based registration
3. Comprehensive conformance testing
4. Performance-oriented schema design

This approach balances flexibility for custom implementations with standardization through OpenFGA's storage abstraction layer.
