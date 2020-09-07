package config

import (
	"os"
)

var (
	port string
	host string = "0.0.0.0"

	databaseUrl string
	githubUrl   string
)

func init() {
	port = os.Getenv("PORT")
	host = os.Getenv("HOST")

	databaseUrl = os.Getenv("DATABASE_URL")
	githubUrl = os.Getenv("GITHUB_URL")
}

func GetHost() string {
	return host
}

func GetPort() string {
	return port
}

func GetDBUrl() string {
	return databaseUrl
}

func GetGithubUrl() string {
	return githubUrl
}
