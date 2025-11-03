package service

import (
	"context"

	"tx-pattern/server/pkg/tx"
)

type Repo interface {
	SomeTransactionalOperation(ctx context.Context, txCtx tx.Context) error
}

type Service struct {
	repo      Repo
	txManager tx.TransactionManager
}

func NewService(repo Repo, txMgr tx.TransactionManager) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) PerformOperation(ctx context.Context) (string, error) {
	var id string

	err := s.txManager.Do(ctx, func(txCtx tx.Context) error {
		return s.repo.SomeTransactionalOperation(ctx, txCtx)
	})

	return id, err
}
