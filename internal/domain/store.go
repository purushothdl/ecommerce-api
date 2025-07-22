package domain

import "context"

type Store interface {
    ExecTx(ctx context.Context, fn func(q *Queries) error) error
}
