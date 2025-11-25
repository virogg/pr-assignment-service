package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	infraerr "github.com/virogg/pr-assignment-service/internal/infrastructure/errors"

	sq "github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRPostgresRepository struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPRPostgresRepository(db *pgxpool.Pool, c *trmpgx.CtxGetter) *PRPostgresRepository {
	return &PRPostgresRepository{
		db:     db,
		getter: c,
	}
}

func (r *PRPostgresRepository) Create(ctx context.Context, pr *entities.PullRequest) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("pull_requests").
		Columns("id", "name", "author_id", "status", "created_at").
		Values(pr.ID, pr.Name, pr.AuthorID, pr.Status.String(), pr.CreatedAt)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `insert PR` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domainerr.ErrPRExists
		}
		return fmt.Errorf("%w: during `create PR`: %w", infraerr.ErrDB, err)
	}

	if len(pr.ReviewerIDs) > 0 {
		if err := r.AddReviewers(ctx, pr.ID, pr.ReviewerIDs); err != nil {
			return err
		}
	}

	return nil
}

func (r *PRPostgresRepository) GetByID(ctx context.Context, id string) (*entities.PullRequest, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get PR` query: %w", err)
	}

	var pr entities.PullRequest

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrPRNotFound
		}
		return nil, fmt.Errorf("%w: during `get PR`: %w", infraerr.ErrDB, err)
	}

	reviewers, err := r.GetReviewers(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.ReviewerIDs = reviewers

	return &pr, nil
}

func (r *PRPostgresRepository) UpdateStatus(ctx context.Context, id string, status vo.PRStatus) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("pull_requests").
		Set("status", status.String()).
		Where(sq.Eq{"id": id})

	// if merge
	if status == vo.PRStatusMerged {
		queryBuilder = queryBuilder.Set("merged_at", sq.Expr("now()"))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `update PR status` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `update PR status`: %w", infraerr.ErrDB, err)
	}

	if result.RowsAffected() == 0 {
		return domainerr.ErrPRNotFound
	}

	return nil
}

func (r *PRPostgresRepository) GetReviewers(ctx context.Context, id string) ([]string, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("reviewer_id").
		From("pr_reviewers").
		Where(sq.Eq{"pull_request_id": id}).
		OrderBy("assigned_at")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get reviewers` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get reviewers`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	reviewerIDs := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("%w: during `scan user`: %w", infraerr.ErrDB, err)
		}
		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get reviewers`: %w", infraerr.ErrDB, err)
	}

	return reviewerIDs, nil
}

func (r *PRPostgresRepository) AddReviewers(ctx context.Context, id string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}

	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("pr_reviewers").
		Columns("pull_request_id", "reviewer_id")

	for _, reviewerID := range reviewerIDs {
		queryBuilder = queryBuilder.Values(id, reviewerID)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `add reviewers` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `add reviewers`: %w", infraerr.ErrDB, err)
	}

	return nil
}

func (r *PRPostgresRepository) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Delete("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID}).
		Where(sq.Eq{"reviewer_id": reviewerID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `remove reviewer` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `remove reviewer`: %w", infraerr.ErrDB, err)
	}

	if result.RowsAffected() == 0 {
		return domainerr.ErrReviewerNotAssigned
	}

	return nil
}

func (r *PRPostgresRepository) GetByReviewer(ctx context.Context, id string) ([]*entities.PullRequest, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("pr.id", "pr.name", "pr.author_id", "pr.status", "pr.created_at", "pr.merged_at").
		From("pull_requests pr").
		Join("pr_reviewers prrev ON pr.id = prrev.pull_request_id").
		Where(sq.Eq{"prrev.reviewer_id": id}).
		OrderBy("pr.created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get PR by reviewer` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get PR by reviewer`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	prs := make([]*entities.PullRequest, 0)
	for rows.Next() {
		var pr entities.PullRequest

		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("%w: during `scan PR`: %w", infraerr.ErrDB, err)
		}

		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get PR by reviewer`: %w", infraerr.ErrDB, err)
	}

	return prs, nil
}

func (r *PRPostgresRepository) GetOpenByReviewers(ctx context.Context, reviewerIDs []string) ([]*entities.PullRequest, error) {
	if len(reviewerIDs) == 0 {
		return []*entities.PullRequest{}, nil
	}

	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("DISTINCT pr.id", "pr.name", "pr.author_id", "pr.status", "pr.created_at", "pr.merged_at").
		From("pull_requests pr").
		Join("pr_reviewers prrev ON pr.id = prrev.pull_request_id").
		Where(sq.Eq{"prrev.reviewer_id": reviewerIDs}).
		Where(sq.Eq{"pr.status": vo.PRStatusOpen.String()}).
		OrderBy("pr.created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get open PRs by reviewers` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get open PRs by reviewers` query`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	prs := make([]*entities.PullRequest, 0)
	for rows.Next() {
		var pr entities.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("%w: during `scan PR`: %w", infraerr.ErrDB, err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get open PRs by reviewers` query`: %w", infraerr.ErrDB, err)
	}

	return prs, nil
}

func (r *PRPostgresRepository) GetAll(ctx context.Context, status *vo.PRStatus, authorID *string) ([]*entities.PullRequest, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests")

	if status != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"status": status.String()})
	}

	if authorID != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"author_id": *authorID})
	}

	queryBuilder.OrderBy("created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get all PRs` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get all PRs`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	prs := make([]*entities.PullRequest, 0)
	for rows.Next() {
		var pr entities.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("%w: during `scan PR`: %w", infraerr.ErrDB, err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get open PRs by reviewers` query`: %w", infraerr.ErrDB, err)
	}

	return prs, nil
}
