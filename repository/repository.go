package repository

import (
	"errors"
	"fmt"

	"github.com/Leantar/fimserver/modules/casbin"
	"github.com/Leantar/fimserver/server"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var errEmptyResultSet = errors.New("result set was empty")

type Config struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"Password"`
	DBName   string `yaml:"db_name"`
	Timezone string `yaml:"timezone"`
}

type PgRepository struct {
	db *sqlx.DB
}

func New(conf Config) *PgRepository {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		conf.Host, conf.User, conf.Password, conf.DBName, conf.Port, conf.Timezone)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("failed to connect to database")
	}

	return &PgRepository{db: db}
}

func (r *PgRepository) ApplySchema() error {
	for _, schema := range schemas {
		_, err := r.db.Exec(schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PgRepository) Endpoints() server.EndpointRepository {
	return &PgEndpointRepository{
		db: r.db,
	}
}

func (r *PgRepository) BaselineFsObjects() server.BaselineFsObjectRepository {
	return &PgBaselineRepository{
		db: r.db,
	}
}

func (r *PgRepository) Alerts() server.AlertRepository {
	return &PgAlertRepository{
		db: r.db,
	}
}

func (r *PgRepository) Rules() casbin.RuleRepository {
	return &PgRuleRepository{
		db: r.db,
	}
}

func (r *PgRepository) IsEmptyResultSetError(err error) bool {
	return errors.Is(err, errEmptyResultSet)
}
