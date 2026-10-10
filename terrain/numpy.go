package terrain

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var shapePattern = regexp.MustCompile(`'shape':\s*\((\d+),\s*(\d+)\s*\)`)

// readNumpy reads a two-dimensional .npy array of little-endian float32 or float64 in C order.
func readNumpy(path string) (values []float32, rows, columns int, err error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}
	header, data, err := splitNumpy(contents)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("%s: %w", path, err)
	}
	if strings.Contains(header, "'fortran_order': True") {
		return nil, 0, 0, fmt.Errorf("%s: Fortran-ordered arrays are not supported", path)
	}
	match := shapePattern.FindStringSubmatch(header)
	if match == nil {
		return nil, 0, 0, fmt.Errorf("%s: expected a two-dimensional array", path)
	}
	rows, _ = strconv.Atoi(match[1])
	columns, _ = strconv.Atoi(match[2])
	values, err = decodeNumbers(header, data, rows*columns)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("%s: %w", path, err)
	}
	return values, rows, columns, nil
}

func splitNumpy(contents []byte) (string, []byte, error) {
	if len(contents) < 10 || string(contents[1:6]) != "NUMPY" {
		return "", nil, fmt.Errorf("not a .npy file")
	}
	headerLength, start := int(binary.LittleEndian.Uint16(contents[8:10])), 10
	if contents[6] >= 2 {
		headerLength, start = int(binary.LittleEndian.Uint32(contents[8:12])), 12
	}
	if start+headerLength > len(contents) {
		return "", nil, fmt.Errorf("truncated header")
	}
	return string(contents[start : start+headerLength]), contents[start+headerLength:], nil
}

func decodeNumbers(header string, data []byte, count int) ([]float32, error) {
	values := make([]float32, count)
	switch {
	case strings.Contains(header, "'<f4'"):
		if len(data) < 4*count {
			return nil, fmt.Errorf("truncated data")
		}
		for index := range values {
			values[index] = math.Float32frombits(binary.LittleEndian.Uint32(data[4*index:]))
		}
	case strings.Contains(header, "'<f8'"):
		if len(data) < 8*count {
			return nil, fmt.Errorf("truncated data")
		}
		for index := range values {
			values[index] = float32(math.Float64frombits(binary.LittleEndian.Uint64(data[8*index:])))
		}
	default:
		return nil, fmt.Errorf("only little-endian float32 or float64 arrays are supported")
	}
	return values, nil
}
