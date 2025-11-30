package validation

import (
	"regexp"
	"strings"
)

// ValidRegions contains all valid Riot API region codes
var ValidRegions = map[string]bool{
	"na":   true, // North America
	"euw":  true, // Europe West
	"eune": true, // Europe Nordic & East
	"kr":   true, // Korea
	"jp":   true, // Japan
	"br":   true, // Brazil
	"lan":  true, // Latin America North
	"las":  true, // Latin America South
	"oce":  true, // Oceania
	"tr":   true, // Turkey
	"ru":   true, // Russia
	"ph":   true, // Philippines
	"sg":   true, // Singapore
	"th":   true, // Thailand
	"tw":   true, // Taiwan
	"vn":   true, // Vietnam
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult holds all validation errors
type ValidationResult struct {
	Errors []ValidationError `json:"errors"`
}

// IsValid returns true if there are no validation errors
func (validationResult *ValidationResult) IsValid() bool {
	return len(validationResult.Errors) == 0
}

// AddError adds a validation error to the result
func (validationResult *ValidationResult) AddError(field string, message string) {
	validationResult.Errors = append(validationResult.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// GetErrorMessages returns all error messages as a single string
func (validationResult *ValidationResult) GetErrorMessages() string {
	messages := make([]string, len(validationResult.Errors))
	for i, validationError := range validationResult.Errors {
		messages[i] = validationError.Field + ": " + validationError.Message
	}
	return strings.Join(messages, "; ")
}

// SummonerRequest represents the request body for summoner lookup
type SummonerRequest struct {
	Region   string `json:"region"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// MatchRequest represents the request body for match history lookup
type MatchRequest struct {
	Region   string `json:"region"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
	PUUID    string `json:"puuid"`
	Count    int    `json:"count"`
}

// AnalyzeRequest represents the request body for player analysis
type AnalyzeRequest struct {
	Region   string `json:"region"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// ValidateSummonerRequest validates a summoner request
func ValidateSummonerRequest(request *SummonerRequest) *ValidationResult {
	result := &ValidationResult{}

	validateRegion(request.Region, result)
	validateGameName(request.GameName, result)
	validateTagLine(request.TagLine, result)

	return result
}

// ValidateMatchRequest validates a match history request
func ValidateMatchRequest(request *MatchRequest) *ValidationResult {
	result := &ValidationResult{}

	validateRegion(request.Region, result)

	// Either PUUID or GameName+TagLine must be provided
	if request.PUUID != "" {
		validatePUUID(request.PUUID, result)
	} else {
		validateGameName(request.GameName, result)
		validateTagLine(request.TagLine, result)
	}

	validateCount(request.Count, result)

	return result
}

// ValidateAnalyzeRequest validates an analyze player request
func ValidateAnalyzeRequest(request *AnalyzeRequest) *ValidationResult {
	result := &ValidationResult{}

	validateRegion(request.Region, result)
	validateGameName(request.GameName, result)
	validateTagLine(request.TagLine, result)

	return result
}

// validateRegion checks if region is valid
func validateRegion(region string, result *ValidationResult) {
	if region == "" {
		result.AddError("region", "region is required")
		return
	}

	lowercaseRegion := strings.ToLower(region)
	if !ValidRegions[lowercaseRegion] {
		result.AddError("region", "invalid region. Valid regions: na, euw, eune, kr, jp, br, lan, las, oce, tr, ru, ph, sg, th, tw, vn")
	}
}

// validateGameName checks if game name is valid
func validateGameName(gameName string, result *ValidationResult) {
	if gameName == "" {
		result.AddError("gameName", "gameName is required")
		return
	}

	// Riot game names must be 3-16 characters
	if len(gameName) < 3 {
		result.AddError("gameName", "gameName must be at least 3 characters")
		return
	}

	if len(gameName) > 16 {
		result.AddError("gameName", "gameName must be at most 16 characters")
		return
	}

	// Game names can only contain letters, numbers, spaces, and underscores
	validGameNamePattern := regexp.MustCompile(`^[a-zA-Z0-9 _]+$`)
	if !validGameNamePattern.MatchString(gameName) {
		result.AddError("gameName", "gameName can only contain letters, numbers, spaces, and underscores")
	}
}

// validateTagLine checks if tag line is valid
func validateTagLine(tagLine string, result *ValidationResult) {
	if tagLine == "" {
		result.AddError("tagLine", "tagLine is required")
		return
	}

	// Riot tag lines must be 3-5 characters
	if len(tagLine) < 3 {
		result.AddError("tagLine", "tagLine must be at least 3 characters")
		return
	}

	if len(tagLine) > 5 {
		result.AddError("tagLine", "tagLine must be at most 5 characters")
		return
	}

	// Tag lines can only contain alphanumeric characters
	validTagLinePattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !validTagLinePattern.MatchString(tagLine) {
		result.AddError("tagLine", "tagLine can only contain letters and numbers")
	}
}

// validatePUUID checks if PUUID format is valid
func validatePUUID(puuid string, result *ValidationResult) {
	if puuid == "" {
		result.AddError("puuid", "puuid is required when not using gameName and tagLine")
		return
	}

	// Riot PUUIDs are 78 characters long (base64 encoded)
	if len(puuid) != 78 {
		result.AddError("puuid", "puuid must be 78 characters")
		return
	}

	// PUUIDs contain alphanumeric characters, hyphens, and underscores
	validPUUIDPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPUUIDPattern.MatchString(puuid) {
		result.AddError("puuid", "puuid contains invalid characters")
	}
}

// validateCount checks if count is within valid range
func validateCount(count int, result *ValidationResult) {
	// Count of 0 is allowed (will use default of 20)
	if count < 0 {
		result.AddError("count", "count cannot be negative")
		return
	}

	// Riot API allows max 100 matches per request
	if count > 100 {
		result.AddError("count", "count cannot exceed 100")
	}
}

// NormalizeRegion converts region to lowercase for consistent API calls
func NormalizeRegion(region string) string {
	return strings.ToLower(region)
}
