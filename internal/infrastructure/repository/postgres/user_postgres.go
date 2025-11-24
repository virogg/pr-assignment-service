package postgres

import (
	"context"
	"errors"
	"fmt"

	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	infraerr "github.com/virogg/pr-assignment-service/internal/infrastructure/errors"

	sq "github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	//"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
)

type UserPostgresRepository struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewUserPostgresRepository(db *pgxpool.Pool, c *trmpgx.CtxGetter) *UserPostgresRepository {
	return &UserPostgresRepository{
		db:     db,
		getter: c,
	}
}

func (r *UserPostgresRepository) Create(ctx context.Context, user *entities.User, teamID int64) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("users").
		Columns("id", "username", "team_id", "is_active").
		Values(user.ID, user.Username, teamID, user.IsActive).
		Suffix(`
			ON CONFLICT (id)
			DO UPDATE SET
				username = EXCLUDED.username,
				team_id = EXCLUDED.team_id,
				is_active = EXCLUDED.is_active,
				updated_at = now()
		`)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `insert users` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `delete courier`: %w", infraerr.ErrDB, err)
	}

	return nil
}

func (r *UserPostgresRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("u.id", "u.username", "t.name", "u.is_active").
		From("users u").
		Join("teams t ON u.team_id = t.id").
		Where(sq.Eq{"u.id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get user` query: %w", err)
	}

	var user entities.User

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err = conn.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: during `get user by id`: %w", infraerr.ErrDB, err)
	}

	return &user, nil
}

func (r *UserPostgresRepository) GetByTeamID(ctx context.Context, id string) ([]*entities.User, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("u.id", "u.username", "t.name", "u.is_active").
		From("users u").
		Join("teams t ON u.team_id = t.id").
		Where(sq.Eq{"u.team_id": id}).
		OrderBy("u.username")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get users by team id` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get users by team id`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("%w: during `scan user`: %w", infraerr.ErrDB, err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get users by team id`: %w", infraerr.ErrDB, err)
	}

	return users, nil
}

func (r *UserPostgresRepository) SetActive(ctx context.Context, id string, isActive bool) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("users").
		Set("is_active", isActive).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `set active` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `set active`: %w", infraerr.ErrDB, err)
	}
	if result.RowsAffected() == 0 {
		//return domainerr.ErrUserNotFound
		return infraerr.ErrNotFound
	}

	return nil
}

func (r *UserPostgresRepository) GetActiveUsersInTeam(ctx context.Context, id int64, excludeUserIDs []string) ([]*entities.User, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("u.id", "u.username", "t.name", "u.is_active").
		From("users u").
		Join("teams t ON u.team_id = t.id").
		Where(sq.Eq{"u.team_id": id}).
		Where(sq.Eq{"u.is_active": true})

	if len(excludeUserIDs) > 0 {
		queryBuilder = queryBuilder.Where(sq.NotEq{"u.id": excludeUserIDs})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build `get active users in team` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get active users in team`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("%w: during `scan user`: %w", infraerr.ErrDB, err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get active users in team`: %w", infraerr.ErrDB, err)
	}

	return users, nil
}

func (r *UserPostgresRepository) SetUsersInactive(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("users").
		Set("is_active", false).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": userIDs})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build `set users inactive` query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: during `set users inactive`: %w", infraerr.ErrDB, err)
	}

	return nil
}
