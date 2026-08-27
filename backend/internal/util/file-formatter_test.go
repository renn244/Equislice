package util

import (
	"errors"
	"testing"
)

func TestFormatFileName(t *testing.T) {
	tests := []struct {
		name     string
		template string
		row      int
		col      int
		want     string
		wantErr  error
	}{
		{"empty template", "", 4, 8, "tile_r4_c8", nil},
		{"handle when {row} does not exist", "slice_{col}_{row", 4, 8, "", errors.New("failed to format file: row is not in the template")},
		{"handle when {col} does not exist", "slice_{row}_{col", 4, 8, "", errors.New("failed to format file: col is not in the template")},
		{"file format working correctly", "slice_{row}_{col}", 4, 8, "slice_4_8", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FormatFileName(test.template, test.row, test.col)

			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("FormatFileName() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf(
						"FormatFileName() expected error = %v, got nil",
						test.wantErr,
					)
				}

				if err.Error() != test.wantErr.Error() {
					t.Errorf(
						"FormatFileName() error = %v, wantErr = %v",
						err,
						test.wantErr,
					)
				}
			}

			if got != test.want {
				t.Errorf(
					"FormatFileName() = %q, want %q",
					got,
					test.want,
				)
			}
		})
	}
}

func TestValidateFileNameFormat(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     bool
		wantErr  error
	}{
		{"empty template", "", true, nil},
		{"only row missing", "slice_{col}_{row", false, errors.New("row is not in the template")},
		{"only col missing", "slice_{row}_{col", false, errors.New("col is not in the template")},
		{"both row and col missing", "slice_{r}_{c", false, errors.New("row and col is not in the template")},
		{"reject file name format with file extension", "slice_{row}_{col}.jpg", false, errors.New("file extension should not be included")},
		{"reject forward slashes format", "slice/{row}/{col}", false, errors.New("slashes should not be included")},
		{"reject backslashes format", "slice\\{row}\\{col}", false, errors.New("slashes should not be included")},
		{"correct file name format", "slice_{row}_{col}", true, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ValidateFileNameFormat(test.template)

			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateFormatFileName() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf(
						"ValidateFormatFileName() expected error = %v, got nil",
						test.wantErr,
					)
				}

				if err.Error() != test.wantErr.Error() {
					t.Errorf(
						"ValidateFormatFileName() error = %v, wantErr = %v",
						err,
						test.wantErr,
					)
				}
			}

			if got != test.want {
				t.Errorf(
					"ValidateFormatFileName() = %v, want %v",
					got,
					test.want,
				)
			}
		})
	}
}

func TestHandleFormatFileNameError(t *testing.T) {
	tests := []struct {
		name       string
		containRow bool
		containCol bool
		want       string
	}{
		{"only row missing", false, true, "row is not in the template"},
		{"only col missing", true, false, "col is not in the template"},
		{"row and col missing", false, false, "row and col is not in the template"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if handleFormatFileNameError(test.containRow, test.containCol) != test.want {
				t.Errorf("expected %v", test.want)
			}
		})
	}
}
