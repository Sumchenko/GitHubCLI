package main

import (
	"errors"
	"flag"
	"fmt"
	"github-user-search/githubapi"
	"os"
)

var ErrNoUsername = errors.New("Флаг username не может быть пустым!")

func validateArgs(username string) error {
	if username == "" {
		return fmt.Errorf("Ошибка проверки аргумента: %w", ErrNoUsername)
	}

	return nil
}

func run() error {
	username := flag.String("username", "", "Name github")

	flag.Parse()

	if err := validateArgs(*username); err != nil {
		return err
	}

	client := githubapi.NewClient()

	body, err := client.FetchUserRepos(*username)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

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
