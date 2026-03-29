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
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationEvent struct{}

type MigrationResponse struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, event MigrationEvent) (MigrationResponse, error) {
	cfg, err := config.LoadConfig[config.MigrationConfig]()
	if err != nil {
		return MigrationResponse{}, err
	}

	var pgPassword = cfg.PGConfig.Password
	if cfg.AWSRegion != "" && cfg.PGPasswordAWSSecretName != "" {
		fmt.Println("Try to load password from AWS Secret Manager")
		pgPassword, err = loadPGPasswordFromAWS(ctx, cfg)
		if err != nil {
			return MigrationResponse{}, err
		}
	}

	query := url.Values{}
	query.Set("sslmode", cfg.PGConfig.SSLMode)
	if cfg.PGConfig.SSLRootCert != "" {
		query.Set("sslrootcert", cfg.PGConfig.SSLRootCert)
	}
	pgDsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.PGConfig.Username, pgPassword),
		Host:     fmt.Sprintf("%s:%d", cfg.PGConfig.Host, cfg.PGConfig.Port),
		Path:     cfg.PGConfig.Database,
		RawQuery: query.Encode(),
	}
	m, err := migrate.New(cfg.Source, pgDsn.String())
	if err != nil {
		return MigrationResponse{}, err
	}

	fmt.Printf(
		"Start migration for DB with user=%s, host=%s, port=%d, database=%s, sslmode=%s, sslrootcert=%s\n",
		cfg.PGConfig.Username,
		cfg.PGConfig.Host,
		cfg.PGConfig.Port,
		cfg.PGConfig.Database,
		cfg.PGConfig.SSLMode,
		cfg.PGConfig.SSLRootCert,
	)

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		return MigrationResponse{}, err
	}

	return MigrationResponse{Message: "Migration completed successfully"}, nil
}

func loadPGPasswordFromAWS(ctx context.Context, cfg *config.MigrationConfig) (string, error) {
	AWScfg, err := aws_config.LoadDefaultConfig(ctx, aws_config.WithRegion(cfg.AWSRegion))
	if err != nil {
		return "", err
	}
	svc := secretsmanager.NewFromConfig(AWScfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(cfg.PGPasswordAWSSecretName),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}
	var secretString string = *result.SecretString

	parsedSecret := make(map[string]string, 2)
	err = json.Unmarshal([]byte(secretString), &parsedSecret)
	if err != nil {
		return "", err
	}

	password, ok := parsedSecret["password"]
	if !ok {
		return "", errors.New("Secret does not cantain 'password' field")
	}

	return password, nil
}

func main() {
	lambda.Start(handler)
}
