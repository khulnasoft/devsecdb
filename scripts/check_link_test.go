package tests

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	// Extract all the links prefix with https://secdb.khulnasoft.com or https://www.secdb.khulnasoft.com in frontend code.
	linkMatcher = regexp.MustCompile(`(?m)https?:\/\/(www\.)?secdb.khulnasoft.com([-a-zA-Z0-9()@:%_\+.~#?&\/\/=]*)`)

	ignores = map[string]bool{
		"node_modules": true,
	}

	extensions = map[string]bool{
		".html": true,
		".js":   true,
		".json": true,
		".ts":   true,
		".vue":  true,
	}

	frontendDirectory = "../frontend"
)

func TestValidateLinks(t *testing.T) {
	links, err := extractLinkRecursive()
	require.NoError(t, err)

	// Check all the links are reachable.
	for link := range links {
		if err := checkLinkWithRetry(link); err != nil {
			require.NoError(t, err)
		}
	}
}

func extractLinkRecursive() (map[string]bool, error) {
	// Initialize the result map.
	links := make(map[string]bool)

	// Define the function to be used with os.Walk.
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if _, ok := ignores[info.Name()]; ok {
			return filepath.SkipDir
		}
		if info.IsDir() || (info.Mode()&os.ModeSymlink) == os.ModeSymlink {
			return nil
		}
		// Check if the file has a valid extension
		if _, ok := extensions[filepath.Ext(info.Name())]; !ok {
			return nil
		}

		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Find all matches using the regular expression
		matches := linkMatcher.FindAllString(string(content), -1)
		// Add matches to the result map
		for _, match := range matches {
			links[match] = true
		}

		return nil
	}

	// Start the recursive traversal using os.Walk
	if err := filepath.Walk(frontendDirectory, walkFn); err != nil {
		return nil, errors.Wrapf(err, "failed to walk directory: %s", frontendDirectory)
	}

	return links, nil
}

func checkLinkWithRetry(link string) error {
	// Request the link and check the response status code is 200.
	resp, err := http.Get(link)
	if err != nil {
		return errors.Wrapf(err, "failed to request link %q", link)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to read link %q", link)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("link %q returned status %q content %q", link, resp.Status, b)
	}
	return nil
}
