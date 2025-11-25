package postgres

import (
	"context"
	"fmt"

	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	infraerr "github.com/virogg/pr-assignment-service/internal/infrastructure/errors"

	sq "github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsPostgresRepository struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewStatsPostgresRepository(db *pgxpool.Pool, c *trmpgx.CtxGetter) *StatsPostgresRepository {
	return &StatsPostgresRepository{
		db:     db,
		getter: c,
	}
}

func (r *StatsPostgresRepository) GetUserStatistics(ctx context.Context) ([]entities.UserStats, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			"u.id",
			"u.username",
			"COUNT(prrev.pull_request_id) AS total_assignments",
			"COUNT(CASE WHEN pr.status = 'OPEN' THEN 1 END) AS open_assignments",
			"COUNT(CASE WHEN pr.status = 'MERGED' THEN 1 END) AS completed_assignments",
		).
		From("users u").
		LeftJoin("pr_reviewers prrev ON u.id = prrev.reviewer_id").
		LeftJoin("pull_requests pr ON prrev.pull_request_id = pr.id").
		GroupBy("u.id", "u.username").
		OrderBy("total_assignments DESC", "u.username")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get user statistics` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get user statistics`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	stats := make([]entities.UserStats, 0)
	for rows.Next() {
		var s entities.UserStats
		if err := rows.Scan(&s.ID, &s.Username, &s.TotalAssignments, &s.OpenAssignments, &s.CompletedAssignments); err != nil {
			return nil, fmt.Errorf("%w: during `scan user stats`: %w", infraerr.ErrDB, err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get team with members`: %w", infraerr.ErrDB, err)
	}

	return stats, nil
}

func (r *StatsPostgresRepository) GetPRStatistics(ctx context.Context) (*entities.PRStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_prs,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) AS open_prs,
			COUNT(CASE WHEN status = 'MERGED' THEN 1 END) AS merged_prs,
			COALESCE(AVG(reviewer_count), 0) AS avg_reviewers_per_pr
		FROM pull_requests pr
		LEFT JOIN (
			SELECT pull_request_id, COUNT(*) AS reviewer_count
			FROM pr_reviewers
			GROUP BY pull_request_id
		) prrev ON pr.id = prrev.pull_request_id
	`

	var stats entities.PRStats

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query).Scan(
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
		&stats.AvgReviewersPerPR,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get PR statistics`: %w", infraerr.ErrDB, err)
	}

	return &stats, nil
}

func (r *StatsPostgresRepository) GetTeamStatistics(ctx context.Context) ([]entities.TeamStats, error) {
	query := `
		SELECT
			t.name,
			COUNT(DISTINCT u.id) AS total_members,
			COUNT(DISTINCT u.id) FILTER (WHERE u.is_active = true) AS active_members,
			COUNT(DISTINCT pr.id) AS total_prs
		FROM teams t
		LEFT JOIN users u ON t.id = u.team_id
		LEFT JOIN pull_requests pr ON u.id = pr.author_id
		GROUP BY t.id, t.name
		ORDER BY t.name
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get team statistics`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	stats := make([]entities.TeamStats, 0)
	for rows.Next() {
		var s entities.TeamStats
		if err := rows.Scan(&s.TeamName, &s.TotalMembers, &s.ActiveMembers, &s.TotalPRs); err != nil {
			return nil, fmt.Errorf("%w: during `scan team stats`: %w", infraerr.ErrDB, err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get team statistics`: %w", infraerr.ErrDB, err)
	}

	return stats, nil
}
