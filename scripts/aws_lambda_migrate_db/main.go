package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"proxy-service/internal/config"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	Up    = "up"
	Down  = "down"
	Steps = "steps"
)

type MigrationEvent struct {
	Command string `json:"command"` // "up" | "down" | "steps", default: "up"
	Steps   int    `json:"steps"`   // only for "steps" command
}

type MigrationResponse struct {
	Message string `json:"message"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event MigrationEvent) (MigrationResponse, error) {
	cfg, err := config.LoadConfig[config.MigrationConfig]()
	if err != nil {
		return MigrationResponse{}, err
	}

	user, err := loadUser(ctx, cfg)
	if err != nil {
		return MigrationResponse{}, err
	}

	query := url.Values{}
	query.Set("sslmode", cfg.PGConfig.SSLMode)
	if cfg.PGConfig.SSLRootCert != "" {
		query.Set("sslrootcert", cfg.PGConfig.SSLRootCert)
	}
	pgDsn := url.URL{
		Scheme:   "postgresql",
		User:     user,
		Host:     fmt.Sprintf("%s:%d", cfg.PGConfig.Host, cfg.PGConfig.Port),
		Path:     cfg.PGConfig.Database,
		RawQuery: query.Encode(),
	}
	m, err := migrate.New(cfg.Source, pgDsn.String())
	if err != nil {
		return MigrationResponse{}, err
	}
	defer m.Close()

	fmt.Printf(
		"Start migration for DB with user=%s, host=%s, port=%d, database=%s, sslmode=%s, sslrootcert=%s\n",
		cfg.PGConfig.Username,
		cfg.PGConfig.Host,
		cfg.PGConfig.Port,
		cfg.PGConfig.Database,
		cfg.PGConfig.SSLMode,
		cfg.PGConfig.SSLRootCert,
	)
	err = runMigration(m, event)
	if err != nil && err.Error() != "no change" {
		return MigrationResponse{}, err
	}

	return MigrationResponse{Message: "Migration completed successfully"}, nil
}

func loadUser(ctx context.Context, cfg *config.MigrationConfig) (*url.Userinfo, error) {
	if cfg.AWSRegion == "" || cfg.PGPasswordAWSSecretName == "" {
		return url.UserPassword(cfg.PGConfig.Username, cfg.PGConfig.Password), nil
	}

	fmt.Println("Load username and password from AWS Secret Manager")
	AWScfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, err
	}
	svc := secretsmanager.NewFromConfig(AWScfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(cfg.PGPasswordAWSSecretName),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		return nil, err
	}
	var secretString string = *result.SecretString

	parsedSecret := make(map[string]string, 2)
	err = json.Unmarshal([]byte(secretString), &parsedSecret)
	if err != nil {
		return nil, err
	}

	username, ok := parsedSecret["username"]
	if !ok {
		return nil, errors.New("Secret does not cantain 'username' field")
	}

	password, ok := parsedSecret["password"]
	if !ok {
		return nil, errors.New("Secret does not cantain 'password' field")
	}

	return url.UserPassword(username, password), nil
}

func runMigration(m *migrate.Migrate, event MigrationEvent) error {
	switch event.Command {
	case "", Up:
		return m.Up()
	case Down:
		return m.Down()
	case Steps:
		return m.Steps(event.Steps)
	default:
		return fmt.Errorf("Unknown command %s", event.Command)
	}
}
