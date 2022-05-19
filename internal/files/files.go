// Package files provides abstractions for working with files.
package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/httputil"
)

// ReadFromLocalOrRemote reads the contents of a file from the local filesystem or from a remote URL.
func ReadFromLocalOrRemote(pathOrURL string, headers map[string]string) (io.ReadCloser, error) {
	if strings.HasPrefix(pathOrURL, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", pathOrURL, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create HTTP request for %q: %w", pathOrURL, err)
		}
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err := httputil.DoWithRetry(req, client)
		if err != nil {
			return nil, fmt.Errorf("failed to make HTTP request to %q: %w", pathOrURL, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP request to %q failed with status %d", pathOrURL, resp.StatusCode)
		}
		return resp.Body, nil
	}
	file, err := os.Open(pathOrURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", pathOrURL, err)
	}
	return file, nil
}
