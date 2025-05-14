package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prkagrawal/cosmos-bk2/graph/model"
)

func IdToUint(idStr string) (uint, error) {
	if idStr == "" {
		return 0, errors.New("ID string cannot be empty")
	}
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format '%s': %w", idStr, err)
	}
	return uint(idUint64), nil
}

func FormatTime(t time.Time) string {
	// Check if the time is the zero value for time.Time.
	// GORM's default for `time.Time` not explicitly set can be "0001-01-01T00:00:00Z".
	// Depending on schema (nullable or not), you might return "" or an error.
	// For non-nullable DateTime, an actual date string is expected.
	if t.IsZero() || t.Unix() < 0 { // Check for zero time or very old default time
		// This case needs careful handling based on whether the GraphQL field is nullable
		// and what a "zero" date means in your application.
		// If the field is non-nullable in GraphQL, returning "" might lead to GQL errors.
		// For now, let's assume zero time means it's not set, and if GraphQL expects a string,
		// we should ideally not reach here for non-nullable fields with zero times.
		// Or, if it's a valid "zero" representation, format it.
		// Let's be strict: if it's a real date, format it. If zero, and GQL field is string, it's tricky.
		// Assuming GORM populates these from DB, they should be valid or zero.
		// Let's return empty string for zero time, and let GQL validation catch if it's an issue for non-nullable.
		return ""
	}
	return t.Format(time.RFC3339)
}

func FormatNullableTime(t *time.Time) *string {
	if t == nil || t.IsZero() || t.Unix() < 0 {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func ParseDateTimeString(dateTimeStr string) (time.Time, error) {
	if dateTimeStr == "" {
		return time.Time{}, errors.New("date string cannot be empty")
	}
	t, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}
	return t, nil
}

func ParseNullableDateTimeString(dateTimeStr *string) (*time.Time, error) {
	if dateTimeStr == nil || *dateTimeStr == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *dateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	return &t, nil
}

func SaveUpload(upload *model.Upload, subDir string) (string, error) {
	if upload == nil {
		return "", nil // No upload provided
	}

	// Create a unique filename to prevent overwrites
	ext := filepath.Ext(upload.Filename)
	uniqueFilename := uuid.New().String() + ext

	// Define the path to save. Ensure 'uploads' directory and 'subDir' exist.
	uploadDirPath := filepath.Join("uploads", subDir)
	if err := os.MkdirAll(uploadDirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}
	filePath := filepath.Join(uploadDirPath, uniqueFilename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file content to the destination file
	if _, err := io.Copy(dst, upload.File); err != nil {
		return "", fmt.Errorf("failed to copy uploaded file: %w", err)
	}

	// Return a URL or path. For local, it might be a relative path.
	// For a web server, you'd return a URL like "/uploads/avatars/filename.jpg"
	// This path needs to be configurable and accessible by your frontend/clients.
	// Example: return "/" + filePath, // if "uploads" is served statically
	return filePath, nil // Returning raw filepath for simplicity here
}

// func (r *userResolver) Applications(ctx context.Context, obj *model.User) ([]*model.Application, error) {
// 	// Authorization: Generally, a user can see their own applications.
// 	// If `obj` is not the currently authenticated user, consider if this data should be public.
// 	// For simplicity, assuming if we have the user object `obj`, we can show their applications.
// 	// The `Me` query would be the typical way a user sees their own.

// 	var applications []*model.Application
// 	// Using GORM's association feature if properly defined, or a direct query:
// 	if err := r.DB.Where("volunteer_id = ?", obj.ID).
// 		Preload("Project"). // Preload related project for context
// 		Order("applied_at DESC").
// 		Find(&applications).Error; err != nil {
// 		fmt.Printf("Error fetching applications for user %d: %v\n", obj.ID, err) // Replace with logger
// 		return nil, fmt.Errorf("failed to fetch applications: %w", err)
// 	}
// 	return applications, nil
// }
// func (r *userResolver) Engagements(ctx context.Context, obj *model.User) ([]*model.Engagement, error) {
// 	// Similar authorization considerations as Applications resolver.
// 	var engagements []*model.Engagement
// 	if err := r.DB.Where("volunteer_id = ?", obj.ID).
// 		Preload("Project").     // Preload related project
// 		Preload("HoursLogged"). // Preload hours logged for this engagement
// 		Order("start_date DESC").
// 		Find(&engagements).Error; err != nil {
// 		fmt.Printf("Error fetching engagements for user %d: %v\n", obj.ID, err) // Replace with logger
// 		return nil, fmt.Errorf("failed to fetch engagements: %w", err)
// 	}
// 	return engagements, nil
// }
