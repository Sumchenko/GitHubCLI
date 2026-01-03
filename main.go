package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github-user-search/githubapi"
	"os"
	"testing"
	"time"
)

func setupTestArgs(args []string) func() {
	oldArgs := os.Args
	oldFlag := flag.CommandLine

	os.Args = append([]string{"mycli"}, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	return func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError error
	}{
		{
			name:          "Missing Username",
			args:          []string{},
			expectedError: ErrNoUsername,
		},
		{
			name:          "Valid Username",
			args:          []string{"-username=test"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer setupTestArgs(tt.args)()

			err := run()

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Ожидалась ошибка %v, но получили %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалась ошибка, но получили %v", err)
				}
			}

		})
	}
}

var ErrNoUsername = errors.New("Флаг username не может быть пустым!")

func validateArgs(username string) error {
	if username == "" {
		return fmt.Errorf("Ошибка проверки аргумента: %w", ErrNoUsername)
	}

	return nil
}

func run() error {
	username := flag.String("username", "", "Name github")
	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for the GitHub API request (e.g., 5s, 1m)")

	flag.Parse()

	if err := validateArgs(*username); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := githubapi.NewClient()

	repos, err := client.FetchUserRepos(ctx, *username)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("API запрос истек по таймауту %s", timeout)
		}
		return err
	}

	fmt.Printf("--- Репозитории для %s (%d всего)   ---\n", *username, len(repos))
	for i, repo := range repos {
		fmt.Printf("%d. %s\n", i+1, repo.Name)
		fmt.Printf("   URL:  %s\n", repo.HTMLUrl)
		fmt.Printf("   Stars: %d | Lang: %s\n", repo.StargazersCount, repo.Language)
		if repo.Description != "" {
			fmt.Printf("   Desc: %s\n", repo.Description)
		}
		fmt.Println("---")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)

		if errors.Is(err, ErrNoUsername) {
			flag.Usage()
			os.Exit(2)
		}

		os.Exit(1)
	}
}
