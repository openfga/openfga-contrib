package authorizer

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(authorizer Authorizer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		storeIDGetter, ok := req.(FGAStoreIDGetter)
		if !ok {
			// no storeID
		}

		storeID := storeIDGetter.GetStoreId()

		authorizeRequest := &AuthorizeRequest{
			Resource: storeID,
		}

		authorizeResp, err := authorizer.Authorize(ctx, authorizeRequest)
		if err != nil {
			return nil, err
		}

		if authorizeResp.Decision == DecisionDeny {
			return nil, status.Error(codes.PermissionDenied, authorizeResp.Reason)
		}

		return handler(ctx, req)
	}
}
