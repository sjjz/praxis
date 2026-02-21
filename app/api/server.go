package api

import (
	"context"
	"fmt"
	"time"

	"praxis/app/lib"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg   lib.Config
	db    *pgxpool.Pool
	store *lib.Store
}

func NewServer(cfg lib.Config) (*Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := lib.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	store := lib.NewStore(pool)
	if err := store.EnsureUser(ctx, cfg.DevUserID); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ensure user: %w", err)
	}

	return &Server{
		cfg:   cfg,
		db:    pool,
		store: store,
	}, nil
}

func (s *Server) Close() {
	s.db.Close()
}
