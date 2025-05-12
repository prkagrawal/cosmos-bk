package graph

import (
	"github.com/prkagrawal/cosmos-bk2/auth"
	"github.com/prkagrawal/cosmos-bk2/database"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// todos []*model.Todo
	DB          *gorm.DB
	AuthService *auth.AuthService
}

func NewResolver(db *gorm.DB) *Resolver {
	return &Resolver{
		DB:          db,
		AuthService: auth.NewAuthService(database.DB),
	}
}
