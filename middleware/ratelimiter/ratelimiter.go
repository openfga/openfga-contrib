package ratelimiter

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(limiter RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		storeIDGetter, ok := req.(FGAStoreIDGetter)
		if !ok {
			// no FGA StoreID field provided in the request
		}

		storeID := storeIDGetter.GetStoreId()

		takeResp, err := limiter.Take(ctx, &TakeRequest{
			Key:        fmt.Sprintf("%s/%s", info.FullMethod, storeID),
			QuotaUnits: 1,
		})
		if err != nil {
			// handle error
		}

		if takeResp.RemainingQuota <= 0 {
			grpc.SetHeader(ctx, metadata.Pairs(
				"X-RateLimit-Limit", strconv.FormatUint(takeResp.LimitQuota, 10),
				"X-RateLimit-Remaining", strconv.FormatUint(takeResp.RemainingQuota, 10),
				"X-RateLimit-Reset", strconv.FormatUint(takeResp.Reset, 10),
			))

			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}
