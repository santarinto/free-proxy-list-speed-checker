package config

type Config struct {
	AppName       string `toml:"app_name"`
	SourceRepoUrl string `toml:"source_repo_url"`
}
