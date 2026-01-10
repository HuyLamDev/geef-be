package projectinfrastructure

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Project represents a project in the database
type Project struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	UserID      string    `db:"user_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ProjectRepository interface {
	SaveProject(project interface{}) error
	FindProjectByID(id string) (interface{}, error)
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sql.DB, logger *logrus.Logger) ProjectRepository {
	sqlxDB := sqlx.NewDb(db, "postgres")
	return &projectRepositoryImpl{
		db:     sqlxDB,
		logger: logger,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type projectRepositoryImpl struct {
	db      *sqlx.DB
	logger  *logrus.Logger
	builder squirrel.StatementBuilderType
}

func (r *projectRepositoryImpl) SaveProject(project interface{}) error {
	r.logger.WithField("project", project).Info("Saving project")

	// Type assert to map for now (since that's what the controller sends)
	projectMap, ok := project.(map[string]interface{})
	if !ok {
		r.logger.Error("SaveProject: invalid project type")
		return sql.ErrTxDone
	}

	// Build INSERT query using squirrel
	query, args, err := r.builder.Insert("projects").
		Columns("id", "name", "description", "user_id", "created_at", "updated_at").
		Values(
			projectMap["id"],
			projectMap["name"],
			projectMap["description"],
			projectMap["user_id"],
			time.Now(),
			time.Now(),
		).
		Suffix("ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, updated_at = EXCLUDED.updated_at").
		ToSql()

	if err != nil {
		r.logger.WithError(err).Error("SaveProject: failed to build query")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("SaveProject: executing query")

	_, err = r.db.Exec(query, args...)
	if err != nil {
		r.logger.WithError(err).Error("SaveProject: failed to execute query")
		return err
	}

	r.logger.Info("SaveProject: project saved successfully")
	return nil
}

func (r *projectRepositoryImpl) FindProjectByID(id string) (interface{}, error) {
	r.logger.WithField("project_id", id).Info("Finding project by ID")

	// Build SELECT query using squirrel
	query, args, err := r.builder.Select("id", "name", "description", "user_id", "created_at", "updated_at").
		From("projects").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		r.logger.WithError(err).Error("FindProjectByID: failed to build query")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("FindProjectByID: executing query")

	var project Project
	err = r.db.Get(&project, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WithField("project_id", id).Debug("FindProjectByID: project not found")
			return nil, nil // Project not found
		}
		r.logger.WithError(err).Error("FindProjectByID: failed to execute query")
		return nil, err
	}

	r.logger.WithField("project_id", id).Debug("FindProjectByID: project found")
	return &project, nil
}