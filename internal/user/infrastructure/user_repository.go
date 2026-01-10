package userinfrastructure

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// User represents a user in the database
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Provider  string    `db:"provider"`
	AvatarURL string    `db:"avatar_url"`
	FullName  string    `db:"full_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserRepository interface {
	SaveUser(user interface{}) error
	FindUserByID(id string) (interface{}, error)
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger *logrus.Logger) UserRepository {
	sqlxDB := sqlx.NewDb(db, "postgres")
	return &userRepositoryImpl{
		db:     sqlxDB,
		logger: logger,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type userRepositoryImpl struct {
	db      *sqlx.DB
	logger  *logrus.Logger
	builder squirrel.StatementBuilderType
}

func (r *userRepositoryImpl) SaveUser(user interface{}) error {
	r.logger.WithField("user", user).Info("Saving user")

	// Type assert to map for now (since that's what the controller sends)
	userMap, ok := user.(map[string]interface{})
	if !ok {
		r.logger.Error("SaveUser: invalid user type")
		return sql.ErrTxDone // Using a generic error, you might want to define custom errors
	}

	// Build INSERT query using squirrel
	query, args, err := r.builder.Insert("users").
		Columns("id", "email", "provider", "avatar_url", "full_name", "created_at", "updated_at").
		Values(
			userMap["sub"],
			userMap["email"],
			userMap["provider"],
			userMap["avatar_url"],
			userMap["full_name"],
			time.Now(),
			time.Now(),
		).
		Suffix("ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, provider = EXCLUDED.provider, avatar_url = EXCLUDED.avatar_url, full_name = EXCLUDED.full_name, updated_at = EXCLUDED.updated_at").
		ToSql()

	if err != nil {
		r.logger.WithError(err).Error("SaveUser: failed to build query")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("SaveUser: executing query")

	_, err = r.db.Exec(query, args...)
	if err != nil {
		r.logger.WithError(err).Error("SaveUser: failed to execute query")
		return err
	}

	r.logger.Info("SaveUser: user saved successfully")
	return nil
}

func (r *userRepositoryImpl) FindUserByID(id string) (interface{}, error) {
	r.logger.WithField("user_id", id).Info("Finding user by ID")

	// Build SELECT query using squirrel
	query, args, err := r.builder.Select("id", "email", "provider", "avatar_url", "full_name", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		r.logger.WithError(err).Error("FindUserByID: failed to build query")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("FindUserByID: executing query")

	var user User
	err = r.db.Get(&user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WithField("user_id", id).Debug("FindUserByID: user not found")
			return nil, nil // User not found
		}
		r.logger.WithError(err).Error("FindUserByID: failed to execute query")
		return nil, err
	}

	r.logger.WithField("user_id", id).Debug("FindUserByID: user found")
	return &user, nil
}