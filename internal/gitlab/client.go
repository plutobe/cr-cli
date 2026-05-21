package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client interacts with the GitLab API.
type Client struct {
	baseURL string
	token   string
}

// NewClient creates a new GitLab API client.
func NewClient(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token}
}

// GetMergeRequestDiff retrieves the diff for a merge request.
func (c *Client) GetMergeRequestDiff(project string, mrID int) (string, error) {
	encodedProject := url.PathEscape(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/changes", c.baseURL, encodedProject, mrID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitLab API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Changes []struct {
			Diff string `json:"diff"`
		} `json:"changes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	var fullDiff string
	for _, change := range result.Changes {
		fullDiff += change.Diff + "\n"
	}
	return fullDiff, nil
}
