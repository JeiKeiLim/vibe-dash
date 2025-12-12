package domain

import (
	"errors"
	"testing"
)

func TestConfidence_String(t *testing.T) {
	tests := []struct {
		name       string
		confidence Confidence
		want       string
	}{
		{"uncertain", ConfidenceUncertain, "Uncertain"},
		{"likely", ConfidenceLikely, "Likely"},
		{"certain", ConfidenceCertain, "Certain"},
		{"invalid negative", Confidence(-1), "Unknown"},
		{"invalid large", Confidence(100), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.confidence.String(); got != tt.want {
				t.Errorf("Confidence.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseConfidence(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Confidence
		wantErr error
	}{
		{"valid lowercase uncertain", "uncertain", ConfidenceUncertain, nil},
		{"valid lowercase likely", "likely", ConfidenceLikely, nil},
		{"valid lowercase certain", "certain", ConfidenceCertain, nil},
		{"valid uppercase", "CERTAIN", ConfidenceCertain, nil},
		{"valid mixed case", "Likely", ConfidenceLikely, nil},
		{"with leading spaces", "  certain", ConfidenceCertain, nil},
		{"with trailing spaces", "certain  ", ConfidenceCertain, nil},
		{"with both spaces", "  certain  ", ConfidenceCertain, nil},
		{"empty string", "", ConfidenceUncertain, nil},
		{"invalid", "invalid", ConfidenceUncertain, ErrInvalidConfidence},
		{"gibberish", "xyz123", ConfidenceUncertain, ErrInvalidConfidence},
		{"partial match", "cert", ConfidenceUncertain, ErrInvalidConfidence},
		{"typo certan", "certan", ConfidenceUncertain, ErrInvalidConfidence},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfidence(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseConfidence() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseConfidence() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ParseConfidence() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseConfidence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfidence_Constants(t *testing.T) {
	// Verify iota values are as expected
	if ConfidenceUncertain != 0 {
		t.Errorf("ConfidenceUncertain = %d, want 0", ConfidenceUncertain)
	}
	if ConfidenceLikely != 1 {
		t.Errorf("ConfidenceLikely = %d, want 1", ConfidenceLikely)
	}
	if ConfidenceCertain != 2 {
		t.Errorf("ConfidenceCertain = %d, want 2", ConfidenceCertain)
	}
}
