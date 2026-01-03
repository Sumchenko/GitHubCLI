package githubapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockServer(t testing.TB, statusCode int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "go-github-cli-v1" {
			t.Errorf("Ожидался User-Agent 'go-github-cli-v1', получен %s", r.Header.Get("User-Agent"))
		}

		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
	return server
}

func TestFetchUserRepos(t *testing.T) {
	tests := []struct {
		name              string
		username          string
		serverStatus      int
		serverBody        string
		expectError       bool
		expectedRepoCount int
	}{
		{
			name:              "Succes Case",
			username:          "testuser",
			serverStatus:      http.StatusOK,
			serverBody:        `[{"name": "repo1", "stargazers_count": 10}, {"name": "repo2", "stargazers_count": 5}]`,
			expectError:       false,
			expectedRepoCount: 2,
		},
		{
			name:              "Not Found (404 )",
			username:          "nonexistent",
			serverStatus:      http.StatusNotFound,
			serverBody:        `{"message": "Not Found"}`,
			expectError:       true,
			expectedRepoCount: 0,
		},
		{
			name:              "Invalid JSON",
			username:          "invalidjson",
			serverStatus:      http.StatusOK,
			serverBody:        `{"name": "repo1"`,
			expectError:       true,
			expectedRepoCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(t, tt.serverStatus, tt.serverBody)
			defer server.Close()

			client := NewClient()
			client.BaseURL = server.URL

			repos, err := client.FetchUserRepos(context.Background(), tt.username)

			if tt.expectError {
				if err == nil {
					t.Fatal("Ожидалась ошибка, но получено nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Не ожидалась ошибка, но получено: %v", err)
				}

				if len(repos) != tt.expectedRepoCount {
					t.Errorf("Ожидалось %d репозиториев, получено %d", tt.expectedRepoCount, len(repos))
				}
			}
		})
	}
}

const sampleJSON = `[
	{"name": "repo1", "html_url": "url1", "description": "desc1", "stargazers_count": 10, "language": "Go"},
	{"name": "repo2", "html_url": "url2", "description": "desc2", "stargazers_count": 5, "language": "Go"}
]`

func BenchmarkJSONUnmarshal(b *testing.B) {
	data := []byte(sampleJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var repos []Repository
		if err := json.Unmarshal(data, &repos); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFetchUserRepos(b *testing.B) {
	server := mockServer(b, http.StatusOK, sampleJSON)
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.FetchUserRepos(ctx, "testuser"); err != nil {
			b.Fatal(err)
		}
	}
}
