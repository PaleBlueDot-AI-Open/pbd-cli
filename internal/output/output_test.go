package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected string
	}{
		{
			name:     "simple map",
			data:     map[string]string{"key": "value"},
			expected: "{\"key\":\"value\"}\n",
		},
		{
			name:     "nested structure",
			data:     map[string]interface{}{"name": "test", "count": 42},
			expected: "{\"count\":42,\"name\":\"test\"}\n",
		},
		{
			name:     "empty map",
			data:     map[string]string{},
			expected: "{}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintJSON(&buf, tt.data)

			if err != nil {
				t.Errorf("PrintJSON() error = %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("PrintJSON() = %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		rows     [][]string
		contains []string
	}{
		{
			name:     "simple table",
			headers:  []string{"Name", "Value"},
			rows:     [][]string{{"item1", "value1"}, {"item2", "value2"}},
			contains: []string{"item1", "value1", "item2", "value2"},
		},
		{
			name:     "empty rows",
			headers:  []string{"Col1", "Col2"},
			rows:     [][]string{},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintTable(&buf, tt.headers, tt.rows)

			output := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("PrintTable() output missing %q", s)
				}
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		unixTs   int64
		expected string
	}{
		{
			name:     "zero timestamp",
			unixTs:   0,
			expected: "-",
		},
		{
			name:     "valid timestamp",
			unixTs:   1609459200, // 2021-01-01 00:00:00 UTC
			expected: "2021-01-01", // Check date portion only (timezone dependent)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.unixTs)
			if tt.unixTs == 0 {
				if result != tt.expected {
					t.Errorf("FormatTime(%d) = %q, want %q", tt.unixTs, result, tt.expected)
				}
			} else {
				if !strings.HasPrefix(result, tt.expected) {
					t.Errorf("FormatTime(%d) = %q, want prefix %q", tt.unixTs, result, tt.expected)
				}
			}
		})
	}
}

func TestFormatQuota(t *testing.T) {
	tests := []struct {
		name      string
		unlimited bool
		quota     int64
		expected  string
	}{
		{
			name:      "unlimited quota",
			unlimited: true,
			quota:     0,
			expected:  "unlimited",
		},
		{
			name:      "limited quota",
			unlimited: false,
			quota:     100000,
			expected:  "100000",
		},
		{
			name:      "zero limited quota",
			unlimited: false,
			quota:     0,
			expected:  "0",
		},
		{
			name:      "large quota",
			unlimited: false,
			quota:     9999999999,
			expected:  "9999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatQuota(tt.unlimited, tt.quota)
			if result != tt.expected {
				t.Errorf("FormatQuota(%v, %d) = %q, want %q", tt.unlimited, tt.quota, result, tt.expected)
			}
		})
	}
}