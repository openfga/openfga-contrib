# RFC: SQL Storage Provider Extensibility

## Status
Draft

## Summary
This RFC describes the extensibility mechanism for implementing custom SQL storage providers in OpenFGA using a two-tier driver pattern combined with the plugin system.

## Motivation
OpenFGA provides a flexible storage abstraction layer that allows custom implementations beyond the built-in PostgreSQL and MySQL support. Organizations may need to:
- Use alternative SQL databases (SQLite, CockroachDB, etc.)
- Implement custom storage optimizations
- Add database-specific features (partitioning, sharding, etc.)
- Support legacy or proprietary database systems

## Extensibility Mechanism

### Architecture

SQL storage providers use a two-tier architecture:

1. **Driver Tier**: Implements `storage.OpenFGADatastoreDriver`
   - Factory for creating datastore instances
   - Handles URI-based configuration via `Open(uri string)` method
   - Returns configured `storage.OpenFGADatastore` instance

2. **Datastore Tier**: Implements `storage.OpenFGADatastore`
   - Executes all storage operations (23 required methods)
   - Manages database connections and lifecycle

### Plugin Registration

Storage providers register with OpenFGA through the plugin system via the `InitPlugin` function:

```go
func InitPlugin(pm *plugin.PluginManager) error {
    driver := &MyDriver{}
    return pm.RegisterOpenFGADatastore("mydb", driver)
}
```

The plugin manager calls `InitPlugin` during startup, and the driver becomes available for use with the registered name.

### Configuration

Storage providers receive configuration through a URI string passed to the `Open` method:

```
mydb://host:port/database?param1=value1&param2=value2
```

The driver is responsible for parsing the URI and initializing the database connection appropriately.

## Implementation Requirements

To implement a custom SQL storage provider:

1. **Implement `storage.OpenFGADatastoreDriver`** with an `Open(uri string)` method
2. **Implement `storage.OpenFGADatastore`** with all 23 required methods covering:
   - Store management (Create, Get, List, Delete)
   - Authorization model operations
   - Tuple operations (Write, Read, queries)
   - Assertion management
   - Change tracking
   - Status and configuration
3. **Create plugin entry point** with `InitPlugin(pm *plugin.PluginManager)` function
4. **Validate with conformance tests** using `github.com/openfga/openfga/pkg/storage/test`

## Testing

Implementations must pass OpenFGA's conformance test suite to ensure compatibility:

```go
func TestOpenFGA_StorageConformance(t *testing.T) {
    ds := New() // Your datastore implementation
    test.RunAllTests(t, ds)
}
```

## Reference Implementation

See `storage/sqlite/sqlite.go` for a complete example structure of a SQL storage provider implementation.

## References

- OpenFGA Storage Interface: `github.com/openfga/openfga/pkg/storage`
- Plugin System: `github.com/openfga/openfga/pkg/plugin`
- Conformance Tests: `github.com/openfga/openfga/pkg/storage/test`
- Reference Implementation: `storage/sqlite/sqlite.go`
