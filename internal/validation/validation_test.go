package validation

import (
	"strings"
	"testing"
)

// TestValidationResult_IsValid tests the IsValid method
func TestValidationResult_IsValid(t *testing.T) {
	result := &ValidationResult{}

	if !result.IsValid() {
		t.Error("Expected empty result to be valid")
	}

	result.AddError("field", "error message")

	if result.IsValid() {
		t.Error("Expected result with errors to be invalid")
	}
}

// TestValidationResult_AddError tests the AddError method
func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{}

	result.AddError("region", "region is required")

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Field != "region" {
		t.Errorf("Expected field 'region', got '%s'", result.Errors[0].Field)
	}

	if result.Errors[0].Message != "region is required" {
		t.Errorf("Expected message 'region is required', got '%s'", result.Errors[0].Message)
	}
}

// TestValidationResult_GetErrorMessages tests the GetErrorMessages method
func TestValidationResult_GetErrorMessages(t *testing.T) {
	result := &ValidationResult{}

	result.AddError("region", "region is required")
	result.AddError("gameName", "gameName is required")

	errorMessages := result.GetErrorMessages()

	if !strings.Contains(errorMessages, "region: region is required") {
		t.Error("Expected error messages to contain region error")
	}

	if !strings.Contains(errorMessages, "gameName: gameName is required") {
		t.Error("Expected error messages to contain gameName error")
	}
}

// TestValidateSummonerRequest_Valid tests valid summoner request
func TestValidateSummonerRequest_Valid(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected valid request, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateSummonerRequest_ValidUppercaseRegion tests region normalization
func TestValidateSummonerRequest_ValidUppercaseRegion(t *testing.T) {
	request := &SummonerRequest{
		Region:   "NA",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected valid request with uppercase region, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateSummonerRequest_MissingRegion tests missing region
func TestValidateSummonerRequest_MissingRegion(t *testing.T) {
	request := &SummonerRequest{
		Region:   "",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for missing region")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}

// TestValidateSummonerRequest_InvalidRegion tests invalid region
func TestValidateSummonerRequest_InvalidRegion(t *testing.T) {
	request := &SummonerRequest{
		Region:   "invalid",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for invalid region")
	}
}

// TestValidateSummonerRequest_AllRegions tests all valid regions
func TestValidateSummonerRequest_AllRegions(t *testing.T) {
	validRegions := []string{"na", "euw", "eune", "kr", "jp", "br", "lan", "las", "oce", "tr", "ru", "ph", "sg", "th", "tw", "vn"}

	for _, region := range validRegions {
		request := &SummonerRequest{
			Region:   region,
			GameName: "TestPlayer",
			TagLine:  "NA1",
		}

		result := ValidateSummonerRequest(request)

		if !result.IsValid() {
			t.Errorf("Expected region '%s' to be valid, got errors: %s", region, result.GetErrorMessages())
		}
	}
}

// TestValidateSummonerRequest_GameNameTooShort tests game name too short
func TestValidateSummonerRequest_GameNameTooShort(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "AB",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for short game name")
	}
}

// TestValidateSummonerRequest_GameNameTooLong tests game name too long
func TestValidateSummonerRequest_GameNameTooLong(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "ThisGameNameIsTooLongForRiot",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for long game name")
	}
}

// TestValidateSummonerRequest_GameNameInvalidChars tests invalid characters in game name
func TestValidateSummonerRequest_GameNameInvalidChars(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "Test@Player!",
		TagLine:  "NA1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for game name with special characters")
	}
}

// TestValidateSummonerRequest_TagLineTooShort tests tag line too short
func TestValidateSummonerRequest_TagLineTooShort(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "AB",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for short tag line")
	}
}

// TestValidateSummonerRequest_TagLineTooLong tests tag line too long
func TestValidateSummonerRequest_TagLineTooLong(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "TOOLONG",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for long tag line")
	}
}

// TestValidateSummonerRequest_TagLineInvalidChars tests invalid characters in tag line
func TestValidateSummonerRequest_TagLineInvalidChars(t *testing.T) {
	request := &SummonerRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA-1",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for tag line with special characters")
	}
}

// TestValidateMatchRequest_ValidWithRiotID tests valid match request with Riot ID
func TestValidateMatchRequest_ValidWithRiotID(t *testing.T) {
	request := &MatchRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
		Count:    10,
	}

	result := ValidateMatchRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected valid request, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateMatchRequest_ValidWithPUUID tests valid match request with PUUID
func TestValidateMatchRequest_ValidWithPUUID(t *testing.T) {
	// Valid 78-character PUUID
	validPUUID := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdef"

	request := &MatchRequest{
		Region: "na",
		PUUID:  validPUUID,
		Count:  10,
	}

	result := ValidateMatchRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected valid request with PUUID, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateMatchRequest_InvalidPUUIDLength tests invalid PUUID length
func TestValidateMatchRequest_InvalidPUUIDLength(t *testing.T) {
	request := &MatchRequest{
		Region: "na",
		PUUID:  "short-puuid",
		Count:  10,
	}

	result := ValidateMatchRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for short PUUID")
	}
}

// TestValidateMatchRequest_NegativeCount tests negative count
func TestValidateMatchRequest_NegativeCount(t *testing.T) {
	request := &MatchRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
		Count:    -1,
	}

	result := ValidateMatchRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for negative count")
	}
}

// TestValidateMatchRequest_CountTooHigh tests count exceeding maximum
func TestValidateMatchRequest_CountTooHigh(t *testing.T) {
	request := &MatchRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
		Count:    101,
	}

	result := ValidateMatchRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for count exceeding 100")
	}
}

// TestValidateMatchRequest_ZeroCountAllowed tests that zero count is allowed (defaults to 20)
func TestValidateMatchRequest_ZeroCountAllowed(t *testing.T) {
	request := &MatchRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
		Count:    0,
	}

	result := ValidateMatchRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected zero count to be valid, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateAnalyzeRequest_Valid tests valid analyze request
func TestValidateAnalyzeRequest_Valid(t *testing.T) {
	request := &AnalyzeRequest{
		Region:   "na",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	result := ValidateAnalyzeRequest(request)

	if !result.IsValid() {
		t.Errorf("Expected valid request, got errors: %s", result.GetErrorMessages())
	}
}

// TestValidateAnalyzeRequest_MissingFields tests missing fields
func TestValidateAnalyzeRequest_MissingFields(t *testing.T) {
	request := &AnalyzeRequest{
		Region:   "",
		GameName: "",
		TagLine:  "",
	}

	result := ValidateAnalyzeRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request for missing fields")
	}

	if len(result.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(result.Errors))
	}
}

// TestNormalizeRegion tests region normalization
func TestNormalizeRegion(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"NA", "na"},
		{"na", "na"},
		{"EUW", "euw"},
		{"Euw", "euw"},
		{"KR", "kr"},
	}

	for _, testCase := range testCases {
		result := NormalizeRegion(testCase.input)

		if result != testCase.expected {
			t.Errorf("Expected '%s', got '%s'", testCase.expected, result)
		}
	}
}

// TestValidRegions tests that ValidRegions map contains expected regions
func TestValidRegions(t *testing.T) {
	expectedRegions := []string{"na", "euw", "eune", "kr", "jp", "br", "lan", "las", "oce", "tr", "ru", "ph", "sg", "th", "tw", "vn"}

	for _, region := range expectedRegions {
		if !ValidRegions[region] {
			t.Errorf("Expected region '%s' to be in ValidRegions", region)
		}
	}

	// Count should match
	if len(ValidRegions) != len(expectedRegions) {
		t.Errorf("Expected %d regions, got %d", len(expectedRegions), len(ValidRegions))
	}
}

// TestValidateSummonerRequest_MultipleErrors tests multiple validation errors
func TestValidateSummonerRequest_MultipleErrors(t *testing.T) {
	request := &SummonerRequest{
		Region:   "invalid",
		GameName: "AB",
		TagLine:  "AB",
	}

	result := ValidateSummonerRequest(request)

	if result.IsValid() {
		t.Error("Expected invalid request")
	}

	// Should have 3 errors: invalid region, short gameName, short tagLine
	if len(result.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d: %s", len(result.Errors), result.GetErrorMessages())
	}
}
