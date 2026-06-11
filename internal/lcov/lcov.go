package lcov

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type File struct {
	SourceFile  string
	LinesFound  int
	LinesHit    int
	CoveragePct float64
}

func Aggregate(content []byte) (float64, error) {
	files, err := Parse(content)
	if err != nil {
		return 0, err
	}

	var linesFound int
	var linesHit int
	for _, file := range files {
		linesFound += file.LinesFound
		linesHit += file.LinesHit
	}
	if linesFound == 0 {
		return 100, nil
	}

	return float64(linesHit) / float64(linesFound) * 100, nil
}

func Parse(content []byte) ([]File, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var files []File
	var current *File

	appendCurrent := func() error {
		if current == nil {
			return nil
		}
		if current.SourceFile == "" {
			return fmt.Errorf("lcov record missing SF line")
		}
		if current.LinesFound == 0 {
			current.CoveragePct = 100
		} else {
			current.CoveragePct = float64(current.LinesHit) / float64(current.LinesFound) * 100
		}
		files = append(files, *current)
		current = nil
		return nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "TN:"):
			continue
		case strings.HasPrefix(line, "SF:"):
			if err := appendCurrent(); err != nil {
				return nil, err
			}
			current = &File{SourceFile: strings.TrimPrefix(line, "SF:")}
		case strings.HasPrefix(line, "LF:"):
			if current == nil {
				return nil, fmt.Errorf("lcov record missing SF before LF")
			}
			value, err := strconv.Atoi(strings.TrimPrefix(line, "LF:"))
			if err != nil {
				return nil, err
			}
			current.LinesFound = value
		case strings.HasPrefix(line, "LH:"):
			if current == nil {
				return nil, fmt.Errorf("lcov record missing SF before LH")
			}
			value, err := strconv.Atoi(strings.TrimPrefix(line, "LH:"))
			if err != nil {
				return nil, err
			}
			current.LinesHit = value
		case line == "end_of_record":
			if err := appendCurrent(); err != nil {
				return nil, err
			}
		default:
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := appendCurrent(); err != nil {
		return nil, err
	}

	return files, nil
}
