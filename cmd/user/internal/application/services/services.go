package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vardius/golog"
	"google.golang.org/grpc"

	authproto "github.com/vardius/go-api-boilerplate/cmd/auth/proto"
	"github.com/vardius/go-api-boilerplate/cmd/user/internal/application/config"
	"github.com/vardius/go-api-boilerplate/cmd/user/internal/domain/user"
	userpersistence "github.com/vardius/go-api-boilerplate/cmd/user/internal/infrastructure/persistence"
	"github.com/vardius/go-api-boilerplate/pkg/auth"
	"github.com/vardius/go-api-boilerplate/pkg/commandbus"
	"github.com/vardius/go-api-boilerplate/pkg/eventbus"
)

type containerFactory func(ctx context.Context, cfg config.Config) (*ServiceContainer, error)

// NewServiceContainer creates new container
var NewServiceContainer containerFactory

type ServiceContainer struct {
	SQL                       *sql.DB
	Logger                    golog.Logger
	CommandBus                commandbus.CommandBus
	EventBus                  eventbus.EventBus
	UserConn                  *grpc.ClientConn
	AuthConn                  *grpc.ClientConn
	UserRepository            user.Repository
	UserPersistenceRepository userpersistence.UserRepository
	AuthClient                authproto.AuthenticationServiceClient
	TokenAuthorizer           auth.TokenAuthorizer
	Authenticator             auth.Authenticator
}

func (c *ServiceContainer) Close() error {
	var wg sync.WaitGroup
	wg.Add(3)

	var errs []error
	go func() {
		defer wg.Done()
		if c.SQL != nil {
			if err := c.SQL.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		if c.UserConn != nil {
			if err := c.UserConn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		if c.AuthConn != nil {
			if err := c.AuthConn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}()

	wg.Wait()

	var closeErr error
	for _, err := range errs {
		if closeErr == nil {
			closeErr = err
		} else {
			closeErr = fmt.Errorf("%v | %v", closeErr, err)
		}
	}

	return closeErr
}
