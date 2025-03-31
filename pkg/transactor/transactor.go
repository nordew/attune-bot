package transactor

import "context"

type Transactor interface {
	Transact(ctx context.Context, fn func(ctx context.Context) error) error
}
