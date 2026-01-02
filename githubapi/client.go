package githubapi

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL:   "https://api.github.com",
		UserAgent: "go-github-cli-v1",
	}
}

func (c *Client) FetchUserRepos(username string) ([]byte, error) {
	url := fmt.Sprintf("%s/users/%s/repos", c.BaseURL, username)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания запроса: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Не удалось выполнить запрос к %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API вернул статут %d для пользователя %s", resp.StatusCode, username)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Не удалось прочитать тело ответа: %w", err)
	}

	return body, nil
}
