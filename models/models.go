package models

import (
	"time"

	"github.com/google/uuid"
)

type GenericModel struct {
	ID        uint      `gorm:"primary_key" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdatableModel struct {
	GenericModel
	UpdatedAt time.Time `json:"updated_at"`
}

type DeletableModel struct {
	UpdatableModel
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

// swagger:model
type PlainProject struct {

	// The id/slug for the project
	// Required: true
	// Min length: 1
	Slug string `gorm:"unique" json:"id" binding:"required" validate:"min=1,regexp=^[a-zA-Z0-9-]*$"`

	// Name of the project
	// Required: false
	Name *string `json:"name"`

	// Description of the project
	// Required: false
	Description *string `json:"description"`

	// Image of the project
	// Required: false
	Image *[]byte `json:"image"`
}

// swagger:model
type Project struct {
	DeletableModel
	PlainProject
	Users         []User `gorm:"many2many:project_users;"`
	ApiKeys       []*ApiKey
	ProgressItems []*ProgressItem
}

// swagger:model
type User struct {
	UpdatableModel
	KratosID uuid.UUID
}

// swagger:model
type ApiKey struct {
	GenericModel
	ProjectID   uint
	Name        *string
	Description *string
	Value       uuid.UUID `gorm:"unique,type:uuid;default:gen_random_uuid()"`
}

// swagger:model
type ProgressItem struct {
	DeletableModel
	ProjectID   uint
	EID         uuid.UUID
	CompletedAt *time.Time
	State       *string
}
