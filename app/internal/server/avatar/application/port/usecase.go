package port

import "context"

// UseCase is the contract for application scenarios.
type UseCase[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}
