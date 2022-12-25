package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	redisdriver "github.com/go-redis/redis/v9"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/udholdenhed/unotes/auth/internal/config"
	"github.com/udholdenhed/unotes/auth/internal/handler/handlergrpc"
	handlerhttp "github.com/udholdenhed/unotes/auth/internal/handler/rest"
	"github.com/udholdenhed/unotes/auth/internal/service"
	"github.com/udholdenhed/unotes/auth/internal/storage"
	"github.com/udholdenhed/unotes/auth/internal/storage/postgres"
	"github.com/udholdenhed/unotes/auth/internal/storage/redis"
	"github.com/udholdenhed/unotes/auth/pkg/utils"
)

func main() {
	InitLogger()

	postgresDB, err := NewPostgreSQL()
	if err != nil {
		log.Fatal().Err(err).Msg("Filed to connect to PostgreSQL.")
	}

	redisDB, err := NewRedis()
	if err != nil {
		log.Fatal().Err(err).Msg("Filed to connect to Redis.")
	}

	repositories := storage.NewRepositoryProvider(
		storage.WithPostgreSQLUserRepository(postgresDB),
		storage.WithRedisRefreshTokenRepository(redisDB),
	)

	services := service.NewService(&service.OAuth2ServiceOptions{
		AccessTokenSecret:      config.C().Auth.AccessTokenSecret,
		RefreshTokenSecret:     config.C().Auth.RefreshTokenSecret,
		AccessTokenExpiresIn:   config.C().Auth.AccessTokenExpiresIn,
		RefreshTokenExpiresIn:  config.C().Auth.RefreshTokenExpiresIn,
		UserRepository:         repositories.UserRepository,
		RefreshTokenRepository: repositories.RefreshTokenRepository,
	})

	addrHTTP := net.JoinHostPort(
		config.C().Auth.HostHTTP,
		config.C().Auth.PortHTTP,
	)
	serverHTTP := handlerhttp.NewHandler(services, &log.Logger).Init(addrHTTP)

	go func() {
		if err := serverHTTP.ListenAndServe(); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
			default:
				log.Fatal().Err(err).Msg("Error occurred while running HTTP server.")
			}
		}
	}()

	addrGRPC := net.JoinHostPort(
		config.C().Auth.HostGRPC,
		config.C().Auth.PortGRPC,
	)
	serverGRPC := handlergrpc.NewHandler(services, &log.Logger).Init(addrGRPC)

	go func() {
		listener, err := net.Listen("tcp", addrGRPC)
		if err != nil {
			log.Fatal().Err(err).Msg("Error occurred while running gRPC server.")
		}

		if err := serverGRPC.Serve(listener); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
			default:
				log.Fatal().Err(err).Msg("Error occurred while running HTTP server.")
			}
		}
	}()

	log.Info().Msg("Server started.")

	<-utils.GracefulShutdown()
	log.Info().Msg("Shutting down...")

	if err := serverHTTP.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("Error occurred on serverHTTP shutting down.")
	}
	serverGRPC.GracefulStop()

	if err := postgresDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error occurred when disconnecting from the PostgreSQL.")
	}

	if err := redisDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error occurred when disconnecting from the Redis.")
	}

	log.Info().Msg("Shutdown completed.")
}

func InitLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC1123,
	})
}

func NewPostgreSQL() (*sqlx.DB, error) {
	return postgres.NewPostgreSQL(context.Background(), &postgres.Config{
		Host:     config.C().PostgreSQL.Host,
		Port:     config.C().PostgreSQL.Port,
		Username: config.C().PostgreSQL.Username,
		Password: config.C().PostgreSQL.Password,
		DBName:   config.C().PostgreSQL.DBName,
		SSLMode:  config.C().PostgreSQL.SSLMode,
	})
}

func NewRedis() (*redisdriver.Client, error) {
	return redis.NewRedis(context.Background(), &redis.Config{
		Addr:     config.C().Redis.Addr,
		Password: config.C().Redis.Password,
		DB:       config.C().Redis.DB,
	})
}