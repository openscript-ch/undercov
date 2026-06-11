package coverage

import (
	"encoding/base64"
	"path"
	"strings"
)

func EncodePath(relPath string) string {
	cleaned := path.Clean(filepathToSlash(relPath))
	cleaned = strings.TrimPrefix(cleaned, "./")
	return base64.RawURLEncoding.EncodeToString([]byte(cleaned))
}

func DecodePath(encoded string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func StoragePath(branch, relPath string) string {
	return path.Join(".undercov", branch, EncodePath(relPath)+".lcov")
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}
