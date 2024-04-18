package ratelimiter

import "context"

type FGAStoreIDGetter interface {
	GetStoreId() string
}

type TakeRequest struct {
	Key        string
	QuotaUnits uint64
}

type TakeResponse struct {
	Key string

	// The total limit/quota for the rate limit key (maps to RateLimit-Limit header)
	LimitQuota uint64

	// The remaining quota for the rate limit key (maps to RateLimit-Remaining header)
	RemainingQuota uint64

	// The reset period for the rate limit key (maps to RateLimit-Reset header)
	Reset uint64
}

type RateLimiter interface {
	// Take attempts to consume the provided number of quota units from
	// the limit quota for the given key.
	Take(ctx context.Context, req *TakeRequest) (*TakeResponse, error)
}
