package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*pgxpool.Config)

func MaxConns(conns int32) Option {
	return func(c *pgxpool.Config) {
		c.MaxConns = conns
	}
}

func MinConns(conns int32) Option {
	return func(c *pgxpool.Config) {
		c.MinConns = conns
	}
}

func MaxConnLifetime(lifetime time.Duration) Option {
	return func(c *pgxpool.Config) {
		c.MaxConnLifetime = lifetime
	}
}

func MaxConnIdleTime(time time.Duration) Option {
	return func(c *pgxpool.Config) {
		c.MaxConnIdleTime = time
	}
}

func HealthCheckPeriod(period time.Duration) Option {
	return func(c *pgxpool.Config) {
		c.HealthCheckPeriod = period
	}
}
