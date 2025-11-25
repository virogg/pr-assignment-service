package postgres

import (
	"context"
	"errors"
	"fmt"

	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	infraerr "github.com/virogg/pr-assignment-service/internal/infrastructure/errors"

	sq "github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
)

type TeamPostgresRepository struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewTeamPostgresRepository(db *pgxpool.Pool, c *trmpgx.CtxGetter) *TeamPostgresRepository {
	return &TeamPostgresRepository{
		db:     db,
		getter: c,
	}
}

func (r *TeamPostgresRepository) Create(ctx context.Context, team *entities.Team) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("teams").
		Columns("name").
		Values(team.Name).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `insert teams` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(&team.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domainerr.ErrTeamExists
		}
		return fmt.Errorf("%w: during `create team`: %w", infraerr.ErrDB, err)
	}

	return nil
}

func (r *TeamPostgresRepository) GetByName(ctx context.Context, name string) (*entities.Team, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name").
		From("teams").
		Where(sq.Eq{"name": name})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get by team name` query: %w", err)
	}

	var team entities.Team

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(&team.ID, &team.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrTeamNotFound
		}
		return nil, fmt.Errorf("%w: during `get by team name`: %w", infraerr.ErrDB, err)
	}

	return &team, nil
}

func (r *TeamPostgresRepository) GetByID(ctx context.Context, id int64) (*entities.Team, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name").
		From("teams").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get by team id` query: %w", err)
	}

	var team entities.Team

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(&team.ID, &team.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrTeamNotFound
		}
		return nil, fmt.Errorf("%w: during `get by team id`: %w", infraerr.ErrDB, err)
	}

	return &team, nil
}

func (r *TeamPostgresRepository) GetWithMembers(ctx context.Context, name string) (*entities.Team, error) {
	team, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "username", "is_active").
		From("users").
		Where(sq.Eq{"team_id": team.ID}).
		OrderBy("username")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get team with members` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get team with members`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			return nil, fmt.Errorf("%w: during `scan user`: %w", infraerr.ErrDB, err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get team with members`: %w", infraerr.ErrDB, err)
	}

	team.Users = users
	return team, nil
}

func (r *TeamPostgresRepository) TeamExists(ctx context.Context, name string) (bool, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("1").
		From("teams").
		Where("name = ?", name).
		Limit(1)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return false, fmt.Errorf("build `team exists` query: %w", err)
	}

	var exists int

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%w: during `team exists`: %w", infraerr.ErrDB, err)
	}
	return true, nil
}
