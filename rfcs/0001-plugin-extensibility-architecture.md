# RFC 0001: Storage Adapter Plugin Architecture for OpenFGA

**Status:** Draft
**Created:** 2025-11-04
**Authors:** OpenFGA Community

## Abstract

This RFC describes the storage adapter plugin architecture for OpenFGA, enabling community contributions of alternative datastore implementations through a standardized plugin system. The architecture leverages Go's plugin mechanism to provide dynamic loading of storage backends while maintaining clear interface contracts and type safety.

## Motivation

OpenFGA's core authorization system requires a persistent storage layer for:
- Authorization stores and metadata
- Authorization models and type definitions
- Relationship tuples
- Change history and assertions

While OpenFGA ships with built-in storage implementations, different deployment environments and use cases may benefit from alternative storage backends. A standardized storage plugin architecture enables:

- **Community Contributions**: Extend OpenFGA with new storage backends without modifying core code
- **Clear Contracts**: Well-defined interfaces ensure compatibility and correctness
- **Dynamic Loading**: Load storage implementations at runtime via configuration
- **Testing and Validation**: Conformance test suite guarantees interface compliance
- **Deployment Flexibility**: Choose the right storage backend for your environment

### Use Cases

- **Embedded Deployments**: SQLite for edge devices or single-tenant scenarios
- **Cloud-Native**: Managed database services (RDS, Cloud SQL, etc.)
- **NoSQL Backends**: Document stores, wide-column stores for specific scale patterns
- **Custom Persistence**: Specialized storage layers for unique requirements
- **Development/Testing**: In-memory or simplified storage for fast iteration

## Design Overview

### Architecture Principles

1. **Interface-Based Design**: Storage plugins implement well-defined interfaces from OpenFGA core
2. **Driver Pattern**: Separation between driver instantiation and datastore operations
3. **Standardized Initialization**: Consistent plugin lifecycle and registration pattern
4. **Conformance Testing**: Comprehensive test suite validates implementation correctness
5. **Shared Libraries**: Compiled as dynamic shared objects (.so/.dylib/.dll)

### Storage Adapter Extension Point

Storage adapters implement OpenFGA's datastore interface, enabling alternative persistence backends while maintaining compatibility with all OpenFGA features.

**Registration Method:**
```go
pm.RegisterOpenFGADatastore(name string, driver storage.OpenFGADatastoreDriver)
```

**Examples:**
- SQLite (reference implementation)
- PostgreSQL alternatives
- MySQL/MariaDB
- NoSQL datastores (MongoDB, DynamoDB, etc.)
- Custom persistence layers
- In-memory stores

## Plugin Lifecycle

### 1. Plugin Development

Storage plugins must implement the standardized initialization function:

```go
func InitPlugin(pm *plugin.PluginManager) error {
    // Register storage driver with the PluginManager
    driver := &MyStorageDriver{}
    return pm.RegisterOpenFGADatastore("mystorage", driver)
}
```

### 2. Plugin Structure

```
plugins/
└── storage/
    ├── sqlite/
    │   └── main.go          # InitPlugin implementation
    ├── mongodb/
    │   └── main.go          # InitPlugin implementation
    └── custom/
        └── main.go          # InitPlugin implementation
```

### 3. Compilation

Plugins are compiled as shared libraries:

```bash
go build -buildmode=plugin -o mystorage.so main.go
```

### 4. Loading and Registration

1. OpenFGA reads plugin configuration at startup
2. Dynamically loads the plugin shared library
3. Calls `InitPlugin` with a `PluginManager` instance
4. Plugin registers its storage driver through the manager

### 5. Integration

1. Registered storage drivers are added to the datastore factory
2. OpenFGA selects the configured driver at runtime
3. Driver's `Open()` method creates datastore instances
4. Datastore handles all persistence operations

## Storage Plugin Specification

### Interface Contract

Storage plugins implement the `OpenFGADatastore` and `OpenFGADatastoreDriver` interfaces from the OpenFGA core package.

#### Driver Interface

The driver interface provides a factory method for creating datastore instances:

```go
type OpenFGADatastoreDriver interface {
    Open(uri string) (OpenFGADatastore, error)
}
```

**Parameters:**
- `uri`: Connection string or configuration URI for the storage backend

**Returns:**
- `OpenFGADatastore`: Configured datastore instance
- `error`: Any error during initialization

#### Datastore Interface

The datastore interface defines all persistence operations. It includes 30+ methods organized into functional categories:

```go
type OpenFGADatastore interface {
    // Store Operations
    CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error)
    DeleteStore(ctx context.Context, id string) error
    GetStore(ctx context.Context, id string) (*openfgav1.Store, error)
    ListStores(ctx context.Context, options ...ListStoresOption) ([]*openfgav1.Store, error)

    // Authorization Model Operations
    WriteAuthorizationModel(ctx context.Context, storeID string, model *openfgav1.AuthorizationModel) error
    ReadAuthorizationModel(ctx context.Context, storeID string, modelID string) (*openfgav1.AuthorizationModel, error)
    ReadAuthorizationModels(ctx context.Context, storeID string, options ...ReadAuthorizationModelsOption) ([]*openfgav1.AuthorizationModel, error)
    FindLatestAuthorizationModel(ctx context.Context, storeID string) (*openfgav1.AuthorizationModel, error)

    // Relationship Tuple Operations
    Write(ctx context.Context, storeID string, deletes []*openfgav1.TupleKey, writes []*openfgav1.TupleKey) error
    Read(ctx context.Context, storeID string, tupleKey *openfgav1.TupleKey) (TupleIterator, error)
    ReadPage(ctx context.Context, storeID string, tupleKey *openfgav1.TupleKey, options ...ReadPageOption) ([]*openfgav1.Tuple, error)
    ReadUserTuple(ctx context.Context, storeID string, tupleKey *openfgav1.TupleKey) (*openfgav1.Tuple, error)
    ReadUsersetTuples(ctx context.Context, storeID string, filter ReadUsersetTuplesFilter) (TupleIterator, error)

    // Change Tracking
    ReadChanges(ctx context.Context, storeID string, filter ReadChangesFilter, options ...ReadChangesOption) ([]*openfgav1.TupleChange, error)

    // Assertions
    WriteAssertions(ctx context.Context, storeID string, modelID string, assertions []*openfgav1.Assertion) error
    ReadAssertions(ctx context.Context, storeID string, modelID string) ([]*openfgav1.Assertion, error)

    // Resource Management
    Close() error
    IsReady(ctx context.Context) (bool, error)
}
```

**Interface Location:** `github.com/openfga/openfga/pkg/storage`

### Operation Categories

#### Store Operations

Manage authorization store metadata and lifecycle:

- **CreateStore**: Create a new authorization store with metadata
- **GetStore**: Retrieve store metadata by ID
- **ListStores**: List all stores with optional filtering/pagination
- **DeleteStore**: Remove a store and all associated data

#### Authorization Model Operations

Manage type definitions and relationship configurations:

- **WriteAuthorizationModel**: Persist a new authorization model version
- **ReadAuthorizationModel**: Retrieve a specific model by ID
- **ReadAuthorizationModels**: List model versions with pagination
- **FindLatestAuthorizationModel**: Get the most recent model for a store

#### Relationship Tuple Operations

Core tuple storage and retrieval:

- **Write**: Atomic write/delete of relationship tuples
- **Read**: Query tuples matching a key pattern (returns iterator)
- **ReadPage**: Paginated tuple queries
- **ReadUserTuple**: Fetch a specific user-object relationship
- **ReadUsersetTuples**: Query tuples representing userset relationships

#### Change Tracking

Maintain change history for synchronization and auditing:

- **ReadChanges**: Retrieve tuple changes since a given point in time
- Supports continuation tokens for pagination
- Enables change data capture (CDC) patterns

#### Assertions

Store test assertions for authorization model validation:

- **WriteAssertions**: Save expected authorization outcomes
- **ReadAssertions**: Retrieve assertions for testing

#### Resource Management

Lifecycle and health check operations:

- **Close**: Clean up resources and connections
- **IsReady**: Health check for readiness probes

### Plugin Implementation Example

#### Plugin Entry Point

```go
// plugins/storage/sqlite/main.go
package main

import (
    "github.com/openfga/openfga-contrib/storage/sqlite"
)

func InitPlugin(pm *plugin.PluginManager) error {
    sqliteDriver := &sqlite.SQLiteDriver{}
    return pm.RegisterOpenFGADatastore("sqlite", sqliteDriver)
}
```

#### Driver Implementation

```go
// storage/sqlite/sqlite.go
package sqlite

import (
    "database/sql"
    "github.com/openfga/openfga/pkg/storage"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteDriver struct{}

func (d *SQLiteDriver) Open(uri string) (storage.OpenFGADatastore, error) {
    db, err := sql.Open("sqlite3", uri)
    if err != nil {
        return nil, fmt.Errorf("failed to open sqlite database: %w", err)
    }

    // Initialize schema
    if err := initializeSchema(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to initialize schema: %w", err)
    }

    return &SQLite{
        db: db,
    }, nil
}
```

#### Datastore Implementation

```go
type SQLite struct {
    db *sql.DB
}

// Store Operations
func (s *SQLite) CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error) {
    query := `INSERT INTO stores (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
    _, err := s.db.ExecContext(ctx, query, store.Id, store.Name, time.Now(), time.Now())
    if err != nil {
        return nil, fmt.Errorf("failed to create store: %w", err)
    }
    return store, nil
}

func (s *SQLite) GetStore(ctx context.Context, id string) (*openfgav1.Store, error) {
    query := `SELECT id, name, created_at, updated_at FROM stores WHERE id = ?`
    row := s.db.QueryRowContext(ctx, query, id)

    var store openfgav1.Store
    err := row.Scan(&store.Id, &store.Name, &store.CreatedAt, &store.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, storage.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get store: %w", err)
    }
    return &store, nil
}

// Implement all remaining OpenFGADatastore interface methods...
// (30+ methods total)

func (s *SQLite) Close() error {
    return s.db.Close()
}

func (s *SQLite) IsReady(ctx context.Context) (bool, error) {
    return s.db.PingContext(ctx) == nil, nil
}
```

### Testing Requirements

#### Conformance Test Suite

Storage implementations **must** pass the OpenFGA conformance test suite to ensure correctness and compatibility:

```go
// storage/sqlite/sqlite_test.go
package sqlite

import (
    "testing"
    "github.com/openfga/openfga/pkg/storage/test"
)

func TestSQLiteDatastore(t *testing.T) {
    // Create test datastore instance
    driver := &SQLiteDriver{}
    datastore, err := driver.Open(":memory:")
    if err != nil {
        t.Fatalf("failed to create test datastore: %v", err)
    }
    defer datastore.Close()

    // Run comprehensive conformance tests
    test.RunAllTests(t, datastore)
}
```

#### What the Conformance Suite Tests

The conformance test suite validates:

1. **Store Lifecycle**: Create, read, list, delete stores
2. **Authorization Models**: Write, read, list, find latest models
3. **Tuple Operations**:
   - Atomic writes and deletes
   - Query patterns and filtering
   - Pagination and iteration
   - Userset tuple handling
4. **Change Tracking**:
   - Change history accuracy
   - Continuation token behavior
   - Temporal ordering
5. **Assertions**: Write and read assertion correctness
6. **Concurrency**: Thread-safe operations
7. **Consistency**: Data integrity across operations
8. **Error Handling**: Proper error types for edge cases
9. **Resource Management**: Connection lifecycle and cleanup

#### Additional Testing Recommendations

Beyond conformance tests, consider:

- **Performance Benchmarks**: Query latency, throughput, resource usage
- **Stress Testing**: High concurrency, large datasets
- **Failure Scenarios**: Network errors, transaction rollbacks, recovery
- **Migration Testing**: Schema upgrades, data migrations
- **Integration Testing**: Full OpenFGA integration tests

## Best Practices

### For Storage Plugin Authors

#### 1. Interface Implementation

- **Complete Implementation**: Implement all 30+ interface methods
- **Error Handling**: Return appropriate error types (use `storage.ErrNotFound`, etc.)
- **Context Respect**: Honor context cancellation and timeouts
- **Thread Safety**: All methods must be safe for concurrent use

#### 2. Testing

- **Pass Conformance Tests**: 100% pass rate required for production use
- **Add Custom Tests**: Test implementation-specific features
- **Performance Testing**: Benchmark critical operations
- **Error Path Testing**: Validate error handling and recovery

#### 3. Schema Management

- **Versioned Schema**: Track schema versions for migrations
- **Initialization**: Create schema automatically on first use
- **Migration Strategy**: Document upgrade paths between versions
- **Indexing**: Create appropriate indexes for query patterns

#### 4. Connection Management

- **Connection Pooling**: Reuse connections efficiently
- **Configuration**: Support pool size, timeout configuration
- **Health Checks**: Implement robust `IsReady()` checks
- **Resource Cleanup**: Properly close connections in `Close()`

#### 5. Transaction Semantics

- **Atomicity**: Tuple writes/deletes must be atomic operations
- **Isolation**: Prevent dirty reads and write conflicts
- **Consistency**: Maintain referential integrity
- **Durability**: Ensure committed data survives failures

#### 6. Performance Optimization

- **Query Optimization**: Use efficient query patterns
- **Bulk Operations**: Optimize batch writes and reads
- **Caching Strategy**: Consider appropriate caching layers
- **Pagination**: Efficient implementation of continuation tokens

#### 7. Documentation

- **Configuration Options**: Document all connection URI parameters
- **Deployment Guide**: Provide deployment best practices
- **Limitations**: Clearly document any limitations or constraints
- **Examples**: Provide complete working examples

#### 8. Error Handling

- **Standard Errors**: Use OpenFGA's standard error types
- **Wrapped Errors**: Wrap underlying errors with context
- **Retryable Errors**: Distinguish transient from permanent failures
- **Logging**: Log errors with sufficient context for debugging

## Configuration

### Plugin Discovery

Storage plugins are configured through OpenFGA's configuration system:

```yaml
datastore:
  engine: sqlite
  plugins:
    - name: sqlite
      path: /usr/local/lib/openfga/plugins/sqlite.so
  uri: file:/var/lib/openfga/db.sqlite?mode=rwc

# Alternative configuration for custom storage
datastore:
  engine: mongodb
  plugins:
    - name: mongodb
      path: /usr/local/lib/openfga/plugins/mongodb.so
  uri: mongodb://localhost:27017/openfga?replicaSet=rs0
```

### Connection URI Format

Each storage implementation defines its own URI format:

#### SQLite Example
```
file:/path/to/database.db?mode=rwc&cache=shared
```

#### PostgreSQL Example
```
postgres://user:password@host:5432/database?sslmode=require&pool_max_conns=25
```

#### MongoDB Example
```
mongodb://host1,host2,host3/database?replicaSet=rs0&w=majority
```

### Environment Variable Support

Support configuration via environment variables:

```bash
OPENFGA_DATASTORE_ENGINE=sqlite
OPENFGA_DATASTORE_URI=file:/var/lib/openfga/db.sqlite
```

## Security Considerations

### 1. Plugin Trust

- **Code Review**: Thoroughly review plugin code before deployment
- **Process Isolation**: Plugins run in the same process with full access
- **Trusted Sources**: Only load plugins from trusted sources
- **Signature Verification**: Consider requiring signed plugins

### 2. Connection Security

- **Encrypted Connections**: Use TLS/SSL for database connections
- **Credential Management**: Never hardcode credentials in plugin code
- **Secrets Management**: Integrate with secrets management systems
- **Least Privilege**: Use database accounts with minimal required permissions

### 3. Input Validation

- **SQL Injection**: Use parameterized queries exclusively
- **NoSQL Injection**: Validate and sanitize all inputs
- **Resource Limits**: Enforce query timeouts and result size limits
- **Schema Validation**: Validate data against expected schemas

### 4. Data Protection

- **Encryption at Rest**: Support transparent data encryption
- **Encryption in Transit**: Always use encrypted connections
- **Access Control**: Implement proper database-level access controls
- **Audit Logging**: Log security-relevant operations

### 5. Denial of Service Prevention

- **Query Timeouts**: Enforce maximum query execution time
- **Connection Limits**: Limit concurrent connections
- **Rate Limiting**: Consider rate limiting at the storage layer
- **Resource Monitoring**: Monitor and alert on resource exhaustion

## Performance Considerations

### 1. Query Optimization

- **Index Strategy**: Create indexes for common query patterns
- **Query Planning**: Analyze query plans for efficiency
- **Selective Queries**: Minimize data scanned per query
- **Pagination**: Use efficient cursor-based pagination

### 2. Connection Pooling

- **Pool Sizing**: Configure appropriate connection pool size
- **Connection Reuse**: Minimize connection creation overhead
- **Idle Connections**: Configure appropriate idle timeouts
- **Health Checks**: Validate connections before use

### 3. Caching Strategies

- **Query Result Caching**: Cache frequently accessed data
- **Cache Invalidation**: Implement proper cache invalidation
- **Cache Warming**: Pre-populate caches on startup
- **Cache Sizing**: Monitor cache hit rates and adjust sizing

### 4. Batching and Bulk Operations

- **Bulk Writes**: Optimize batch tuple writes
- **Bulk Reads**: Minimize round trips for multiple reads
- **Transaction Batching**: Group operations into transactions
- **Async Processing**: Use async operations where appropriate

### 5. Monitoring and Observability

- **Metrics**: Expose storage operation metrics
- **Latency Tracking**: Track P50, P95, P99 latencies
- **Error Rates**: Monitor and alert on error rates
- **Resource Usage**: Track connection count, memory usage

### 6. Scalability Patterns

- **Read Replicas**: Support read replica configurations
- **Sharding**: Document sharding strategies if applicable
- **Partitioning**: Use table partitioning for large datasets
- **Horizontal Scaling**: Enable horizontal scaling where possible

## Compatibility and Versioning

### 1. Interface Stability

- **Semantic Versioning**: OpenFGA interfaces follow semver
- **Breaking Changes**: Major version increments indicate breaking changes
- **Deprecation Policy**: Deprecated methods maintained for one major version
- **Migration Guides**: Provided for major version upgrades

### 2. Plugin Compatibility

- **Version Requirements**: Document compatible OpenFGA versions
- **Dependency Pinning**: Pin OpenFGA version in go.mod
- **Testing Matrix**: Test against multiple OpenFGA versions
- **Compatibility Badges**: Display supported versions

### 3. Schema Versioning

- **Schema Version Tracking**: Store schema version in database
- **Migration Scripts**: Provide migration scripts for upgrades
- **Backward Compatibility**: Support reading older schema versions
- **Forward Compatibility**: Design schemas for future extensibility

## Example: Complete SQLite Plugin

The reference SQLite implementation demonstrates all concepts:

### Repository Structure
```
storage/sqlite/
├── sqlite.go              # Driver and datastore implementation
├── sqlite_test.go         # Conformance tests
├── schema.sql             # Database schema
└── migrations/            # Schema migration scripts
    ├── 001_initial.sql
    └── 002_add_indexes.sql

plugins/storage/sqlite/
└── main.go                # Plugin entry point
```

### Key Features

- **Complete Interface**: All 30+ methods implemented
- **SQLite Optimizations**: Appropriate indexes and pragmas
- **Transaction Support**: Atomic tuple operations
- **Change Tracking**: Efficient change log implementation
- **Testing**: 100% conformance test pass rate
- **Documentation**: Inline documentation and examples

### Usage Example

```bash
# Build the plugin
cd plugins/storage/sqlite
go build -buildmode=plugin -o sqlite.so main.go

# Configure OpenFGA
cat > config.yaml <<EOF
datastore:
  engine: sqlite
  plugins:
    - name: sqlite
      path: ./sqlite.so
  uri: file:openfga.db?cache=shared&mode=rwc
EOF

# Run OpenFGA with SQLite plugin
openfga run --config config.yaml
```

## References

- **OpenFGA Core**: https://github.com/openfga/openfga
- **Storage Interface**: https://github.com/openfga/openfga/tree/main/pkg/storage
- **Go Plugins**: https://pkg.go.dev/plugin
- **SQLite Plugin**: `storage/sqlite/` in this repository

## Future Enhancements

### 1. Plugin Discovery and Distribution

- **Plugin Registry**: Central registry for discovering community plugins
- **Automated Testing**: CI/CD pipeline for plugin validation
- **Version Management**: Automatic compatibility checking
- **Plugin Marketplace**: Curated collection of verified plugins

### 2. Advanced Features

- **Hot Reloading**: Runtime plugin updates without restart
- **Plugin Sandboxing**: Isolation mechanisms for untrusted plugins
- **Multi-Tenancy**: Storage-level tenant isolation
- **Streaming Operations**: Support for streaming large result sets

### 3. Observability Enhancements

- **Distributed Tracing**: OpenTelemetry integration
- **Structured Logging**: Standardized logging interfaces
- **Metrics Framework**: Common metrics collection interface
- **Health Dashboards**: Pre-built monitoring dashboards

### 4. Performance Improvements

- **Query Caching**: Built-in query result caching layer
- **Connection Multiplexing**: Advanced connection management
- **Async Operations**: Non-blocking operation support
- **Read/Write Splitting**: Automatic read replica routing

### 5. Developer Experience

- **Code Generation**: Generate boilerplate implementation code
- **Testing Framework**: Enhanced testing utilities
- **Migration Tools**: Automated schema migration tooling
- **Documentation Generator**: Auto-generate plugin documentation

## Appendix: File Reference

Key implementation files in this repository:

### Storage Implementation
- `storage/sqlite/sqlite.go`: Complete SQLite datastore implementation
- `storage/sqlite/sqlite_test.go`: Conformance test suite integration

### Plugin Entry Points
- `plugins/storage/sqlite/main.go`: SQLite plugin initialization

### Documentation
- `README.md`: Project overview and getting started guide

## Changelog

- **2025-11-04**: Initial draft focusing on storage adapter plugins
