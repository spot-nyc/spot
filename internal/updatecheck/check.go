package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// githubReleasesAPI is the default base URL used by Check. Tests override it
// via checkWithBaseURL.
const githubReleasesAPI = "https://api.github.com"

// Check queries the GitHub Releases API for the latest published version of
// spot-nyc/spot. Returns the latest tag ("v0.2.0"), whether it is strictly
// newer than currentVersion, and any network or parse error. The context
// controls the overall deadline — callers should pass a short timeout (2s
// recommended).
func Check(ctx context.Context, currentVersion string) (latest string, available bool, err error) {
	return checkWithBaseURL(ctx, currentVersion, githubReleasesAPI)
}

func checkWithBaseURL(ctx context.Context, currentVersion, baseURL string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/repos/spot-nyc/spot/releases/latest", nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false, fmt.Errorf("github api: status %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", false, err
	}
	return body.TagName, isNewer(currentVersion, body.TagName), nil
}
