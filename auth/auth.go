package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
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

// GenerateJWT needs to be consistent. If it stores ID as uint, the middleware/GetuserFromContext must handle it.
// The default jwt.MapClaims will likely store uint as float64.
func GenerateJWT(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID, // user.ID is uint. This will likely be encoded as a number (float64) by JWT library.
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GetUserFromContext retrieves the user from the context.
// It now expects the value associated with userContextKey.
func GetUserFromContext(ctx context.Context) (*model.User, error) {
	userCtxValue := ctx.Value(userContextKey)
	if userCtxValue == nil {
		return nil, errors.New("not authenticated")
	}

	token, ok := userCtxValue.(*jwt.Token)
	if !ok {
		// This should not happen if the middleware is set up correctly.
		return nil, errors.New("invalid token or user type in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims type")
	}

	// userIDClaim, ok := claims["sub"].(string)
	// if !ok {
	// 	return nil, errors.New("userID (sub) claim missing or not a string in token")
	// }
	// fmt.Printf("User ID from token claims: %s\n", userIDClaim)
	// In schema, ID is string, but gorm.Model.ID is uint.
	// GenerateJWT uses user.ID (which is uint). If it's stored as string in JWT, fine.
	// If stored as number in JWT, need to handle that.
	// Assuming user.ID is string or can be converted.
	// If user.ID is uint and your JWT sub is user.ID.String(), then this is okay.
	// But GenerateJWT uses user.ID which is uint. The jwt library probably marshals it as float64.
	// Let's adjust to handle that common case if `sub` is a number in JWT.

	// --- THIS IS THE CRITICAL SECTION ---
	subClaimValue, subClaimExists := claims["sub"]
	if !subClaimExists {
		return nil, errors.New("userID (sub) claim missing in token")
	}

	// For debugging the type of subClaimValue:
	// fmt.Printf("GetUserFromContext: subClaimValue type: %T, value: %v\n", subClaimValue, subClaimValue)

	var userIDString string
	switch subValTyped := subClaimValue.(type) {
	case float64: // JWT standard numeric types often become float64
		userIDString = fmt.Sprintf("%.0f", subValTyped) // Convert float64 to string without decimals
	case string:
		userIDString = subValTyped
	default:
		// This error will give you the exact type if it's neither float64 nor string
		return nil, fmt.Errorf("unexpected type for userID (sub) claim: %T. Value: '%[1]v'", subClaimValue)
	}
	// fmt.Printf("GetUserFromContext: userID (sub) claim value: %s\n", userIDString)

	var user model.User
	// GORM's First method expects the primary key type.
	// If model.User.ID is `uint`, then need to convert `userID` string to `uint`.

	idUint, err := strconv.ParseUint(userIDString, 10, 0) // 0 means match uint size
	if err != nil {
		return nil, fmt.Errorf("could not parse userID from token for DB lookup: %w", err)
	}

	// Preload associations as needed by your `me` query resolvers (e.g., Skills, Causes)
	if err := database.DB.Preload("Skills").Preload("Causes").First(&user, idUint).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user ID %d (from token sub claim '%s') not found in database", idUint, userIDString)
		}
		return nil, fmt.Errorf("failed to retrieve user from database: %w", err)
	}

	// if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
	// 	return nil, fmt.Errorf("user not found: %w", err)
	// }

	return &user, nil
}
