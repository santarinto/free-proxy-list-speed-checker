package config

type Config struct {
	AppName       string `toml:"app_name"`
	SourceRepoUrl string `toml:"source_repo_url"`

	ProxyCollectionList ProxyCollectionList `toml:"proxy_collection_list"`
	Options             Options             `toml:"options"`
}

type ProxyCollectionList struct {
	Socks5 string `toml:"socks5"`
}

type Options struct {
	CacheDir string `toml:"cache_dir"`
}
