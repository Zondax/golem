package zvalidator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "test@sub.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "test+label@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			email:   "test123@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots in local part",
			email:   "test.user@example.com",
			wantErr: false,
		},

		// Basic validation errors
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name:    "missing @",
			email:   "testexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "multiple @",
			email:   "test@test@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing domain",
			email:   "test@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},

		// Length validation errors
		{
			name:    "email too long",
			email:   strings.Repeat("a", 60) + "@" + strings.Repeat("b", 200) + ".com",
			wantErr: true,
			errMsg:  "email exceeds maximum length",
		},

		// Domain validation (business constraint)
		{
			name:    "no dot in domain",
			email:   "test@localhost",
			wantErr: true,
			errMsg:  "email domain must contain at least one dot",
		},

		// Go's mail.ParseAddress will catch these
		{
			name:    "invalid characters",
			email:   "test<script>@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "space in email",
			email:   "test user@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "consecutive dots in local part",
			email:   "test..test@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "dot at start of local part",
			email:   ".test@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "dot at end of local part",
			email:   "test.@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},

		// Security test cases - Go's parser handles these
		{
			name:    "SQL injection attempt",
			email:   "test@example.com' OR '1'='1",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "XSS attempt",
			email:   "<script>alert('xss')</script>@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},

		// Test ParseAddress normalization detection
		{
			name:    "email with display name gets normalized",
			email:   "Test User <test@example.com>",
			wantErr: true,
			errMsg:  "email format was modified during parsing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test edge cases and boundary conditions
func TestValidateEmail_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "exactly max length",
			email:   strings.Repeat("a", 240) + "@example.com", // 254 total
			wantErr: false,
		},
		{
			name:    "one char over max length",
			email:   strings.Repeat("a", 243) + "@example.com", // 255 total (243 + 12 = 255)
			wantErr: true,
			errMsg:  "email exceeds maximum length",
		},
		{
			name:    "valid international domain",
			email:   "test@example.org",
			wantErr: false,
		},
		{
			name:    "valid with hyphen in domain",
			email:   "test@sub-domain.example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
