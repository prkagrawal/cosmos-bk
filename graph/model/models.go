package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"time"

	"gorm.io/gorm"
)

// Custom type for weekday array
type Weekdays []Weekday

func (w *Weekdays) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid weekday array")
	}
	return json.Unmarshal(bytes, w)
}

func (w Weekdays) Value() (driver.Value, error) {
	return json.Marshal(w)
}

type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    string
	Role         UserRole `gorm:"type:varchar(20);index"`
	Bio          string
	LinkedInURL  string
	PortfolioURL string
	Skills       []Skill      `gorm:"many2many:user_skills;"`
	Causes       []Cause      `gorm:"many2many:user_causes;"`
	Availability Availability `gorm:"embedded"`

	// Relationships
	Applications []Application `gorm:"foreignKey:VolunteerID"`
	Engagements  []Engagement  `gorm:"foreignKey:VolunteerID"`
}

type Nonprofit struct {
	gorm.Model
	Name        string
	Description string
	LogoURL     string
	Website     string
	EIN         string        `gorm:"uniqueIndex"`
	Verified    bool          `gorm:"index"`
	Size        NonprofitSize `gorm:"type:varchar(20)"`
	Causes      []Cause       `gorm:"many2many:nonprofit_causes;"`
	Location    Location      `gorm:"embedded"`
	Members     []User        `gorm:"many2many:nonprofit_members;"`
	Projects    []Project     `gorm:"foreignKey:NonprofitID"`
}

type Project struct {
	gorm.Model
	Title        string
	Description  string
	SkillsNeeded []Skill `gorm:"many2many:project_skills;"`

	TimeCommitment TimeCommitment `gorm:"type:varchar(30)"`
	Urgency        UrgencyLevel   `gorm:"type:varchar(10)"`
	Status         ProjectStatus  `gorm:"type:varchar(20);index"`
	StartDate      *time.Time
	EndDate        *time.Time
	NonprofitID    uint

	// Relationships
	Applications []Application `gorm:"foreignKey:ProjectID"`
	Engagements  []Engagement  `gorm:"foreignKey:ProjectID"`
}

type Availability struct {
	HoursPerWeek  int      `gorm:"type:int"`
	DaysAvailable Weekdays `gorm:"type:jsonb"`
	Timezone      string   `gorm:"type:varchar(50)"`
}

type Location struct {
	City    string `gorm:"type:varchar(100)"`
	State   string `gorm:"type:varchar(100)"`
	Country string `gorm:"type:varchar(100)"`
	Remote  bool
}

type Application struct {
	gorm.Model
	Message   *string
	Status    ApplicationStatus `gorm:"type:varchar(20);index"`
	AppliedAt time.Time
	DecidedAt *time.Time

	// Relationships
	VolunteerID uint
	Volunteer   User `gorm:"foreignKey:VolunteerID"`

	ProjectID uint
	Project   Project `gorm:"foreignKey:ProjectID"`
}

type Engagement struct {
	gorm.Model
	StartDate           time.Time
	EndDate             *time.Time
	Status              EngagementStatus `gorm:"type:varchar(20);index"`
	Feedback            *string
	FeedbackSubmittedAt *time.Time

	// Relationships
	VolunteerID uint
	Volunteer   User `gorm:"foreignKey:VolunteerID"`

	ProjectID uint
	Project   Project `gorm:"foreignKey:ProjectID"`

	HoursLogged []HoursLogged `gorm:"foreignKey:EngagementID"`
}

type HoursLogged struct {
	gorm.Model
	Date        time.Time
	Hours       float64
	Description *string
	Approved    *bool
	ApprovedAt  *time.Time

	// Relationships
	EngagementID uint
	Engagement   Engagement `gorm:"foreignKey:EngagementID"`

	ApprovedByID *uint
	ApprovedBy   User `gorm:"foreignKey:ApprovedByID"`
}

type Skill struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex"`
	Category string
}

type Cause struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex"`
	Description string
}

// Upload represents a file upload in GraphQL.
type Upload struct {
	Filename string
	File     io.Reader
}
