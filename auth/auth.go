package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prkagrawal/cosmos-bk2/database"
	"github.com/prkagrawal/cosmos-bk2/graph/model"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{DB: db}
}

func (a *AuthService) CreateUser(ctx context.Context, input model.SignupInput) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := model.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         input.Role,
	}

	result := a.DB.Create(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("user creation failed: %w", result.Error)
	}

	return &user, nil
}

func (a *AuthService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	var user model.User
	if err := a.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func GenerateJWT(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GetUserFromContext(ctx context.Context) (*model.User, error) {
	userCtx := ctx.Value("user")
	if userCtx == nil {
		return nil, errors.New("not authenticated")
	}

	token := userCtx.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	var user model.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}
