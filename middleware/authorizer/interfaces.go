package authorizer

import "context"

type FGAStoreIDGetter interface {
	GetStoreId() string
}

type Decision int

const (
	// DecisionDeny means that an authorizer decided to deny the action.
	DecisionDeny Decision = iota

	// DecisionAllow means that an authorizer decided to allow the action.
	DecisionAllow

	// DecisionNoOpionion means that an authorizer has no opinion on whether
	// to allow or deny an action.
	DecisionNoOpinion
)

type Authorizer interface {
	Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error)
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
