package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/eduardojabes/CodeArena/internal/pkg/entity"
	pb "github.com/eduardojabes/CodeArena/proto/user"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotExists     = errors.New("user not exists")
	ErrInvalidPassword   = errors.New("incorrect user name or password")
)

type UserRepository interface {
	GetUser(ctx context.Context, ID uuid.UUID) (*entity.User, error)
	AddUser(ctx context.Context, user entity.User) error
	SearchUserByUsername(ctx context.Context, name string) (*entity.User, error)
}

type Hasher interface {
	GenerateFromPassword(password string) (string, error)
	CompareHashAndPassword(hashedPassword string, password string) (bool, error)
}

type UserService struct {
	pb.UnimplementedUserServer
	repository UserRepository
	hasher     Hasher
}

func (s *UserService) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := s.repository.SearchUserByUsername(ctx, in.GetName())
	if err != nil {
		err = fmt.Errorf("error while creating user: %w", err)
		return nil, err
	}

	if user != nil {
		return nil, ErrUserAlreadyExists
	}

	password, err := s.hasher.GenerateFromPassword(in.GetPassword())

	if err != nil {
		err = fmt.Errorf("error while encrypting the password: %w", err)

		return nil, err
	}

	newUser := entity.User{
		ID:       uuid.New(),
		Username: in.GetName(),
		Password: password,
	}

	err = s.repository.AddUser(ctx, newUser)

	if err != nil {
		err = fmt.Errorf("error while attempt to access user DB: %w", err)

		return nil, err
	}

	return &pb.CreateUserResponse{UserID: uuid.NewString()}, nil
}

func (s *UserService) GetUserByUserName(ctx context.Context, in *pb.GetUserByUsernameRequest) (*pb.GetUserByUsernameResponse, error) {
	user, err := s.repository.SearchUserByUsername(ctx, in.GetName())
	if err != nil {
		err = fmt.Errorf("error while getting user: %w", err)
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotExists
	}

	return &pb.GetUserByUsernameResponse{UserID: user.ID.String(),
		Name:     user.Username,
		Password: user.Password,
	}, nil
}

func (s *UserService) GetUserByUserNameAndPassword(ctx context.Context, in *pb.GetUserByUserNameAndPasswordRequest) (*pb.GetUserByUserNameAndPasswordResponse, error) {
	user, err := s.GetUserByUserName(ctx, &pb.GetUserByUsernameRequest{
		Name: in.GetName(),
	})

	if err != nil {
		if errors.Is(err, ErrUserNotExists) {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}

	ok, err := s.hasher.CompareHashAndPassword(user.GetPassword(), in.GetPassword())
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrInvalidPassword
	}

	return &pb.GetUserByUserNameAndPasswordResponse{UserID: user.GetUserID()}, nil
}

func (s *UserService) Register(sr grpc.ServiceRegistrar) {
	pb.RegisterUserServer(sr, s)
}

func NewUserService(repository UserRepository, hasher Hasher) *UserService {
	return &UserService{repository: repository, hasher: hasher}
}
