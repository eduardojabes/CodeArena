package user

import (
	"context"
	"errors"
	"testing"

	"github.com/eduardojabes/CodeArena/internal/pkg/entity"
	"github.com/eduardojabes/CodeArena/internal/pkg/util"
	"github.com/eduardojabes/CodeArena/internal/test"
	pb "github.com/eduardojabes/CodeArena/proto/user"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func TestHashPassword(t *testing.T) {
	t.Run("Creating User", func(t *testing.T) {

	})
}

func TestCreateUser(t *testing.T) {
	t.Run("Creating User effectively", func(t *testing.T) {
		input := &pb.CreateUserRequest{Name: "name", Password: "password"}

		repository := &test.MockUserRepository{
			SearchUserByUsernameMock: func(ctx context.Context, name string) (*entity.User, error) {
				return nil, nil
			},
			AddUserMock: func(ctx context.Context, user entity.User) error {
				return nil
			},
		}
		service := NewUserService(repository, util.NewBCryptHasher())

		got, err := service.CreateUser(context.Background(), input)

		if err != nil {
			t.Errorf("error creating user")
		}
		if got.UserID == uuid.Nil.String() {
			t.Errorf("got %v want an ID", got)
		}
	})

	t.Run("Error to create User in SearchUser", func(t *testing.T) {
		want := errors.New("error")
		input := &pb.CreateUserRequest{Name: "name", Password: "password"}

		repository := &test.MockUserRepository{
			SearchUserByUsernameMock: func(ctx context.Context, name string) (*entity.User, error) {
				return nil, want
			},
		}
		service := NewUserService(repository, util.NewBCryptHasher())

		_, got := service.CreateUser(context.Background(), input)

		if !errors.Is(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Error to create User in SearchUser", func(t *testing.T) {
		want := errors.New("error")
		input := &pb.CreateUserRequest{Name: "name", Password: "password"}

		repository := &test.MockUserRepository{
			SearchUserByUsernameMock: func(ctx context.Context, name string) (*entity.User, error) {
				return nil, nil
			},
			AddUserMock: func(ctx context.Context, user entity.User) error {
				return want
			},
		}
		service := NewUserService(repository, util.NewBCryptHasher())

		_, got := service.CreateUser(context.Background(), input)

		if !errors.Is(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Error to create User that exists", func(t *testing.T) {
		input := &pb.CreateUserRequest{Name: "name", Password: "password"}

		user := &entity.User{
			ID:       uuid.New(),
			Username: "user",
			Password: "password",
		}

		repository := &test.MockUserRepository{
			SearchUserByUsernameMock: func(ctx context.Context, name string) (*entity.User, error) {
				return user, nil
			},
		}
		service := NewUserService(repository, util.NewBCryptHasher())

		_, got := service.CreateUser(context.Background(), input)

		if !errors.Is(got, ErrUserAlreadyExists) {
			t.Errorf("got %v want %v", got, ErrUserAlreadyExists)
		}
	})

	t.Run("Error while hashing password", func(t *testing.T) {
		want := errors.New("error")

		input := &pb.CreateUserRequest{Name: "name", Password: "password"}

		hasher := &test.MockHasher{
			GenerateFromPasswordMock: func(password string) (string, error) {
				return "", want
			},
		}

		repository := &test.MockUserRepository{
			SearchUserByUsernameMock: func(ctx context.Context, name string) (*entity.User, error) {
				return nil, nil
			},
		}
		service := NewUserService(repository, hasher)

		_, got := service.CreateUser(context.Background(), input)

		if !errors.Is(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})
}

func TestRegisterUserService(t *testing.T) {
	s := grpc.NewServer()
	service := NewUserService(&test.MockUserRepository{}, util.NewBCryptHasher())

	service.Register(s)
}