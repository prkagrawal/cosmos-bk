package model

type Weekday string
type UserRole string
type ProjectStatus string
type ApplicationStatus string
type EngagementStatus string
type NonprofitSize string
type TimeCommitment string
type UrgencyLevel string

const (
	Monday    Weekday = "MONDAY"
	Tuesday   Weekday = "TUESDAY"
	Wednesday Weekday = "WEDNESDAY"
	Thursday  Weekday = "THURSDAY"
	Friday    Weekday = "FRIDAY"
	Saturday  Weekday = "SATURDAY"
	Sunday    Weekday = "SUNDAY"
)

const (
	Volunteer      UserRole = "VOLUNTEER"
	NonprofitAdmin UserRole = "NONPROFIT_ADMIN"
	PlatformAdmin  UserRole = "PLATFORM_ADMIN"
)

const (
	Draft         ProjectStatus = "DRAFT"
	PendingReview ProjectStatus = "PENDING_REVIEW"
	Active        ProjectStatus = "ACTIVE"
	InProgress    ProjectStatus = "IN_PROGRESS"
	Completed     ProjectStatus = "COMPLETED"
	Cancelled     ProjectStatus = "CANCELLED"
)

const (
	Pending   ApplicationStatus = "PENDING"
	Accepted  ApplicationStatus = "ACCEPTED"
	Rejected  ApplicationStatus = "REJECTED"
	Withdrawn ApplicationStatus = "WITHDRAWN"
)

const (
	EngagementActive    EngagementStatus = "ACTIVE"
	EngagementCompleted EngagementStatus = "COMPLETED"
	EngagementCancelled EngagementStatus = "CANCELLED"
)

const (
	Small     NonprofitSize = "SMALL"
	Medium    NonprofitSize = "MEDIUM"
	Large     NonprofitSize = "LARGE"
	VeryLarge NonprofitSize = "VERY_LARGE"
)

const (
	LessThan5Hours      TimeCommitment = "LESS_THAN_5_HOURS"
	FiveToTenHours      TimeCommitment = "FIVE_TO_TEN_HOURS"
	TenToTwentyHours    TimeCommitment = "TEN_TO_TWENTY_HOURS"
	MoreThanTwentyHours TimeCommitment = "MORE_THAN_TWENTY_HOURS"
)

const (
	Low      UrgencyLevel = "LOW"
	Mid      UrgencyLevel = "MEDIUM"
	High     UrgencyLevel = "HIGH"
	Critical UrgencyLevel = "CRITICAL"
)
