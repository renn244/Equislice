package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func FormatFileName(template string, row int, col int) (string, error) {
	if template == "" {
		return fmt.Sprintf("tile_r%d_c%d", row, col), nil
	}

	containRow := strings.Contains(template, "{row}")
	containCol := strings.Contains(template, "{col}")
	if !containRow || !containCol {
		errMessage := handleFormatFileNameError(containRow, containCol)
		return "", fmt.Errorf("failed to format file: %w", errors.New(errMessage))
	}

	colReplace := strings.ReplaceAll(template, "{col}", strconv.Itoa(col))
	rowReplace := strings.ReplaceAll(colReplace, "{row}", strconv.Itoa(row))

	return rowReplace, nil
}

func ValidateFileNameFormat(template string) (bool, error) {
	if template == "" {
		return true, nil
	}

	containsFileExtensions := strings.Contains(template, ".")
	if containsFileExtensions {
		return false, errors.New("file extension should not be included")
	}

	containsSlashes := strings.Contains(template, "/") || strings.Contains(template, "\\")
	if containsSlashes {
		return false, errors.New("slashes should not be included")
	}

	containRow := strings.Contains(template, "{row}")
	containCol := strings.Contains(template, "{col}")
	if !containRow || !containCol {
		errMessage := handleFormatFileNameError(containRow, containCol)
		return false, errors.New(errMessage)
	}

	return true, nil
}

func handleFormatFileNameError(containRow bool, containCol bool) string {
	var errMessage string

	if !containRow && !containCol {
		errMessage = "row and col is not in the template"
	} else if containRow == false {
		errMessage = "row is not in the template"
	} else if containCol == false {
		errMessage = "col is not in the template"
	}

	return errMessage
}
