package main

import (
	"context"
	"math"

	"google.golang.org/grpc"

	"github.com/openfga/openfga-contrib/middleware/authorizer"
	"github.com/openfga/openfga-contrib/middleware/ratelimiter"
	"github.com/openfga/openfga/pkg/plugin"
)

type mockAuthorizer struct{}

func (m *mockAuthorizer) Authorize(
	ctx context.Context,
	req *authorizer.AuthorizeRequest,
) (*authorizer.AuthorizeResponse, error) {
	if req.Resource == "01HSNVPXRVE6MGW080RPZ4C880" {
		return &authorizer.AuthorizeResponse{
			Decision: authorizer.DecisionAllow,
		}, nil
	}

	return &authorizer.AuthorizeResponse{
		Decision: authorizer.DecisionDeny,
	}, nil
}

type mockLimiter struct{}

func (m *mockLimiter) Take(
	ctx context.Context,
	req *ratelimiter.TakeRequest,
) (*ratelimiter.TakeResponse, error) {
	return &ratelimiter.TakeResponse{
		LimitQuota:     10,
		RemainingQuota: 0,
		Reset:          math.MaxUint64,
	}, nil
}

func InitPlugin(pm *plugin.PluginManager) error {
	// 1. Load structured middleware config, for example 'authorizer'

	// 2. Initialize middleware implementations based on config
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		authorizer.UnaryServerInterceptor(&mockAuthorizer{}),
		ratelimiter.UnaryServerInterceptor(&mockLimiter{}),
	}

	// 3. Register them with the OpenFGA runtime
	return pm.RegisterUnaryServerInterceptors(unaryInterceptors...)
}
