# RFC 0001: Plugin Extensibility Architecture for OpenFGA

**Status:** Draft
**Created:** 2025-11-04
**Authors:** OpenFGA Community

## Abstract

This RFC describes the plugin-based extensibility architecture for OpenFGA, enabling community contributions of middleware interceptors and storage implementations through a standardized plugin system. The architecture leverages Go's plugin mechanism to provide dynamic loading of extensions while maintaining clear interface contracts and type safety.

## Motivation

OpenFGA's core authorization system benefits from extensibility in two key areas:

1. **Middleware/Interceptors**: Cross-cutting concerns such as authorization, rate limiting, logging, metrics, and auditing
2. **Storage Implementations**: Alternative datastore backends beyond the built-in options

A standardized plugin architecture enables:
- Community contributions without modifying core OpenFGA code
- Clear contracts and interfaces for extensions
- Dynamic loading and configuration of extensions
- Separation of core functionality from optional features
- Easier testing and validation of extensions

## Design Overview

### Architecture Principles

1. **Interface-Based Design**: All extensions implement well-defined interfaces
2. **Standardized Initialization**: Consistent plugin lifecycle and registration
3. **Native gRPC Integration**: Middleware leverages standard gRPC interceptor patterns
4. **Driver Pattern**: Storage abstractions separate driver instantiation from datastore operations
5. **Shared Libraries**: Compiled as dynamic shared objects (.so/.dylib/.dll)

### Extension Points

The architecture provides two primary extension points:

#### 1. Middleware Interceptors

Middleware extensions integrate as gRPC `UnaryServerInterceptor` instances, providing cross-cutting functionality for all RPC calls.

**Registration Method:**
```go
pm.RegisterUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor)
```

**Examples:**
- Authorization checking (FGA-on-FGA)
- Rate limiting
- Request logging
- Metrics collection
- Audit trails

#### 2. Storage Datastores

Storage extensions implement OpenFGA's datastore interface, enabling alternative persistence backends.

**Registration Method:**
```go
pm.RegisterOpenFGADatastore(name string, driver storage.OpenFGADatastoreDriver)
```

**Examples:**
- SQLite (reference implementation)
- Alternative SQL databases
- NoSQL datastores
- Custom persistence layers

## Plugin Lifecycle

### 1. Plugin Development

Plugins must implement the standardized initialization function:

```go
func InitPlugin(pm *plugin.PluginManager) error {
    // Register components with the PluginManager
    return nil
}
```

### 2. Plugin Structure

```
plugins/
├── middleware/
│   └── main.go          # InitPlugin implementation
└── storage/
    └── sqlite/
        └── main.go      # InitPlugin implementation
```

### 3. Compilation

Plugins are compiled as shared libraries:

```bash
go build -buildmode=plugin -o plugin.so main.go
```

### 4. Loading and Registration

OpenFGA dynamically loads the plugin and calls `InitPlugin` with a `PluginManager` instance. The plugin registers its components through the provided manager.

### 5. Integration

Registered components are integrated into the appropriate OpenFGA subsystems:
- Middleware interceptors are added to the gRPC server chain
- Storage drivers are registered in the datastore factory

## Middleware Plugin Specification

### Interface Contract

Middleware plugins create implementations of domain-specific interfaces and wrap them in gRPC interceptors.

#### Authorizer Interface

```go
type Authorizer interface {
    Authorize(ctx context.Context, request *AuthorizeRequest) (*AuthorizeResponse, error)
}

type AuthorizeRequest struct {
    Resource   string
    Permission string
    Subject    string
}

type AuthorizeResponse struct {
    Decision Decision
    Reason   string
}

type Decision int32

const (
    DecisionNoOpinion Decision = iota
    DecisionAllow
    DecisionDeny
)
```

**File:** `middleware/authorizer/interfaces.go`

#### RateLimiter Interface

```go
type RateLimiter interface {
    Take(ctx context.Context, request *TakeRequest) (*TakeResponse, error)
}

type TakeRequest struct {
    Key        string
    QuotaUnits int64
}

type TakeResponse struct {
    LimitQuota     int64
    RemainingQuota int64
    Reset          int64  // Unix timestamp
}
```

**File:** `middleware/ratelimiter/interfaces.go`

### Interceptor Creation

The middleware package provides helper functions to create gRPC interceptors from interface implementations:

```go
func UnaryServerInterceptor(a Authorizer) grpc.UnaryServerInterceptor
func UnaryServerInterceptor(rl RateLimiter) grpc.UnaryServerInterceptor
```

### Plugin Implementation Example

```go
// plugins/middleware/main.go
package main

import (
    "github.com/openfga/openfga-contrib/middleware/authorizer"
    "github.com/openfga/openfga-contrib/middleware/ratelimiter"
    "google.golang.org/grpc"
)

func InitPlugin(pm *plugin.PluginManager) error {
    unaryInterceptors := []grpc.UnaryServerInterceptor{
        authorizer.UnaryServerInterceptor(&customAuthorizer{}),
        ratelimiter.UnaryServerInterceptor(&customRateLimiter{}),
    }
    return pm.RegisterUnaryServerInterceptors(unaryInterceptors...)
}

type customAuthorizer struct {}
type customRateLimiter struct {}

// Implement respective interfaces...
```

### Request Context Integration

Middleware can access request-specific information through interface detection:

```go
type FGAStoreIDGetter interface {
    GetStoreId() string
}

if storeGetter, ok := req.(FGAStoreIDGetter); ok {
    storeID := storeGetter.GetStoreId()
    // Use storeID for authorization/rate limiting
}
```

### Error Handling

Middleware interceptors should return appropriate gRPC status codes:

- **Authorization Denial**: `codes.PermissionDenied`
- **Rate Limit Exceeded**: `codes.ResourceExhausted`
- **Internal Errors**: `codes.Internal`

### Response Headers

Middleware can set response headers for client communication:

```go
grpc.SetHeader(ctx, metadata.Pairs(
    "X-RateLimit-Limit", fmt.Sprint(resp.LimitQuota),
    "X-RateLimit-Remaining", fmt.Sprint(resp.RemainingQuota),
    "X-RateLimit-Reset", fmt.Sprint(resp.Reset),
))
```

## Storage Plugin Specification

### Interface Contract

Storage plugins implement the `OpenFGADatastore` and `OpenFGADatastoreDriver` interfaces from the OpenFGA core.

#### Driver Interface

```go
type OpenFGADatastoreDriver interface {
    Open(uri string) (OpenFGADatastore, error)
}
```

#### Datastore Interface (Partial)

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
    // ... 30+ methods total
}
```

**File:** Defined in `github.com/openfga/openfga/pkg/storage`

### Plugin Implementation Example

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

### Driver Implementation

```go
// storage/sqlite/sqlite.go
package sqlite

type SQLiteDriver struct{}

func (d *SQLiteDriver) Open(uri string) (storage.OpenFGADatastore, error) {
    db, err := sql.Open("sqlite3", uri)
    if err != nil {
        return nil, err
    }
    return &SQLite{db: db}, nil
}

type SQLite struct {
    db *sql.DB
}

// Implement all 30+ OpenFGADatastore methods...
```

### Testing Requirements

Storage implementations must pass the OpenFGA conformance test suite:

```go
// storage/sqlite/sqlite_test.go
package sqlite

import (
    "testing"
    "github.com/openfga/openfga/pkg/storage/test"
)

func TestSQLiteDatastore(t *testing.T) {
    datastore, err := NewSQLiteDatastore(":memory:")
    if err != nil {
        t.Fatal(err)
    }

    test.RunAllTests(t, datastore)
}
```

The conformance test suite validates:
- Store lifecycle operations
- Authorization model CRUD
- Tuple operations (write, read, delete)
- Query operations
- Change tracking
- Assertions
- Concurrency and consistency

## Best Practices

### For Plugin Authors

1. **Implement Complete Interfaces**: Ensure all interface methods are implemented
2. **Pass Conformance Tests**: Storage plugins must pass the full test suite
3. **Handle Errors Gracefully**: Return appropriate error types and status codes
4. **Document Configuration**: Clearly document any required configuration
5. **Version Dependencies**: Pin OpenFGA version dependencies in go.mod
6. **Thread Safety**: Ensure implementations are safe for concurrent use
7. **Resource Cleanup**: Implement proper resource cleanup (Close methods, etc.)
8. **Performance Considerations**: Profile and optimize critical paths

### For Middleware Authors

1. **Minimal Performance Impact**: Keep interceptors lightweight
2. **Context Propagation**: Properly propagate context and cancellation
3. **Logging**: Use structured logging for debugging
4. **Graceful Degradation**: Handle missing optional interfaces
5. **Idempotency**: Ensure interceptors can be called multiple times safely
6. **Observability**: Expose metrics for monitoring

### For Storage Authors

1. **Transaction Support**: Implement proper transaction semantics where applicable
2. **Indexing**: Create appropriate indexes for query performance
3. **Schema Management**: Document database schema and migration strategy
4. **Connection Pooling**: Implement efficient connection management
5. **Backup and Recovery**: Document backup procedures
6. **Data Integrity**: Ensure consistency and durability guarantees

## Configuration

### Plugin Discovery

Plugins are discovered through configuration. Example configuration format:

```yaml
plugins:
  - type: middleware
    path: /path/to/middleware-plugin.so
  - type: storage
    name: sqlite
    path: /path/to/sqlite-plugin.so
    config:
      uri: /var/lib/openfga/data.db
```

### Plugin Manager Configuration

The PluginManager accepts configuration for:
- Plugin search paths
- Load order (for middleware chain ordering)
- Plugin-specific configuration parameters

## Security Considerations

1. **Plugin Trust**: Plugins run in the same process as OpenFGA and have full access. Only load trusted plugins.
2. **Input Validation**: Plugins must validate all inputs to prevent injection attacks
3. **Secret Management**: Handle credentials and secrets securely
4. **Audit Logging**: Middleware should support audit trails for security-critical operations
5. **Isolation**: Future versions may explore sandboxing mechanisms

## Performance Considerations

1. **Middleware Overhead**: Each interceptor adds latency. Minimize processing time in the hot path.
2. **Storage Efficiency**: Optimize query patterns and indexing strategies
3. **Connection Pooling**: Reuse connections and resources effectively
4. **Caching**: Consider caching strategies for frequently accessed data
5. **Monitoring**: Expose metrics for performance monitoring and tuning

## Compatibility and Versioning

1. **Interface Stability**: Core interfaces follow semantic versioning
2. **Breaking Changes**: Major version bumps indicate breaking interface changes
3. **Deprecation Policy**: Deprecated interfaces maintained for one major version
4. **Testing**: Plugins should test against multiple OpenFGA versions
5. **Dependencies**: Pin compatible versions in go.mod

## Examples

### Complete Middleware Plugin

See `plugins/middleware/main.go` for a complete example implementing:
- Custom authorizer with FGA-on-FGA pattern
- Rate limiter with per-store quotas

### Complete Storage Plugin

See `storage/sqlite/` for a complete SQLite implementation including:
- Full OpenFGADatastore interface implementation
- Driver pattern with connection management
- Conformance test suite integration

## References

- OpenFGA Core: https://github.com/openfga/openfga
- gRPC Interceptors: https://github.com/grpc/grpc-go
- Go Plugins: https://pkg.go.dev/plugin

## Future Enhancements

1. **Plugin Sandboxing**: Explore isolation mechanisms for untrusted plugins
2. **Hot Reloading**: Support runtime plugin updates without restart
3. **Plugin Registry**: Central registry for discovering community plugins
4. **Observability Hooks**: Standardized interfaces for metrics and tracing
5. **Stream Interceptors**: Support for streaming RPC interceptors
6. **Configuration Validation**: Schema validation for plugin configurations
7. **Dependency Management**: Plugin dependency resolution and compatibility checking

## Appendix: File Reference

Key implementation files in this repository:

- `middleware/authorizer/interfaces.go`: Authorizer interface definition
- `middleware/authorizer/authorizer.go`: Authorizer interceptor implementation
- `middleware/ratelimiter/interfaces.go`: RateLimiter interface definition
- `middleware/ratelimiter/ratelimiter.go`: RateLimiter interceptor implementation
- `storage/sqlite/sqlite.go`: SQLite storage implementation
- `storage/sqlite/sqlite_test.go`: Conformance test suite
- `plugins/middleware/main.go`: Example middleware plugin
- `plugins/storage/sqlite/main.go`: Example storage plugin

## Changelog

- **2025-11-04**: Initial draft created
