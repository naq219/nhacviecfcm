package middleware

import (
	"errors"
	"regexp"
	"strings"
)

// SQL validation patterns
var (
	dangerousPatterns = []string{
		`(?i)\bDROP\b`,
		`(?i)\bTRUNCATE\b`,
		`(?i)\bALTER\b`,
		`(?i)\bCREATE\b`,
		`(?i)--`,
		`(?i)/\*`,
		`(?i)\*/`,
		`(?i)\bEXEC\b`,
		`(?i)\bEVAL\b`,
		`(?i)\bSYSTEM\b`,
		`(?i)\bSHUTDOWN\b`,
		`(?i)\bGRANT\b`,
		`(?i)\bREVOKE\b`,
		`(?i)\bLOAD_FILE\b`,
		`(?i)\bINTO\s+OUTFILE\b`,
		`(?i)\bINTO\s+DUMPFILE\b`,
	}

	updateWherePattern = regexp.MustCompile(`(?i)UPDATE\s+.*\s+WHERE\s+`)
	deleteWherePattern = regexp.MustCompile(`(?i)DELETE\s+FROM\s+.*\s+WHERE\s+`)
)

// ValidateQuery validates a SQL query for safety
func ValidateQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return errors.New("query cannot be empty")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, query); matched {
			return errors.New("query contains potentially dangerous operations")
		}
	}

	return nil
}

// ValidateSelectQuery validates a SELECT query
func ValidateSelectQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return errors.New("only SELECT statements are allowed")
	}

	return nil
}

// ValidateInsertQuery validates an INSERT query
func ValidateInsertQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "INSERT") {
		return errors.New("only INSERT statements are allowed")
	}

	return nil
}

// ValidateUpdateQuery validates an UPDATE query
func ValidateUpdateQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "UPDATE") {
		return errors.New("only UPDATE statements are allowed")
	}

	// Require WHERE clause for UPDATE
	if !updateWherePattern.MatchString(query) {
		return errors.New("UPDATE statements must include a WHERE clause for safety")
	}

	return nil
}

// ValidateDeleteQuery validates a DELETE query
func ValidateDeleteQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "DELETE") {
		return errors.New("only DELETE statements are allowed")
	}

	// Require WHERE clause for DELETE
	if !deleteWherePattern.MatchString(query) {
		return errors.New("DELETE statements must include a WHERE clause for safety")
	}

	return nil
}
