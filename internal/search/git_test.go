package search

import (
	"testing"
	"time"
)

func TestParseTimeFlag(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input   string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			input:   "1 week ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -7)
				if diff := got.Sub(expected); diff > time.Hour || diff < -time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		{
			input:   "3 days ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -3)
				if diff := got.Sub(expected); diff > time.Hour || diff < -time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		{
			input:   "2 months ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, -2, 0)
				if diff := got.Sub(expected); diff > 48*time.Hour || diff < -48*time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		{
			input:   "1 year ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(-1, 0, 0)
				if diff := got.Sub(expected); diff > 48*time.Hour || diff < -48*time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		{
			input:   "5 weeks ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -35)
				if diff := got.Sub(expected); diff > time.Hour || diff < -time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		{
			input:   "1 day ago",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -1)
				if diff := got.Sub(expected); diff > time.Hour || diff < -time.Hour {
					t.Errorf("expected ~%v, got %v (diff %v)", expected, got, diff)
				}
			},
		},
		// Invalid inputs
		{input: "invalid", wantErr: true},
		{input: "1 ago", wantErr: true},
		{input: "abc days ago", wantErr: true},
		{input: "1 decades ago", wantErr: true},
		{input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTimeFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTimeFlag(%q) should return error", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseTimeFlag(%q) unexpected error: %v", tt.input, err)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestBuildContextContent(t *testing.T) {
	lines := []string{"line1", "line2", "line3", "line4", "line5"}

	// Match at middle (line3, index 2), context=1
	content := buildContextContent(lines, 2, 1)
	if content == "" {
		t.Error("buildContextContent should return non-empty string")
	}
	// Should contain line2, > line3, line4
	if len(content) == 0 {
		t.Error("content should not be empty")
	}

	// Match at start (line1, index 0), context=2
	content = buildContextContent(lines, 0, 2)
	if len(content) == 0 {
		t.Error("content should not be empty for start match")
	}

	// Match at end (line5, index 4), context=2
	content = buildContextContent(lines, 4, 2)
	if len(content) == 0 {
		t.Error("content should not be empty for end match")
	}

	// Single line
	content = buildContextContent([]string{"only"}, 0, 3)
	if content != "> only" {
		t.Errorf("single line context should be '> only', got %q", content)
	}
}
