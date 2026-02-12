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

type ConfigPath struct {
	AppName                  string                   `toml:"app_name"`
	SourceRepoUrl            string                   `toml:"source_repo_url"`
	ProxyCollectionListPatch ProxyCollectionListPatch `toml:"proxy_collection_list"`
	OptionsPatch             OptionsPatch             `toml:"options"`
}

type ProxyCollectionListPatch struct {
	Socks5 *string `toml:"socks5"`
}

type OptionsPatch struct {
	CacheDir *string `toml:"cache_dir"`
}

func (c *Config) ApplyPatch(p ConfigPath) {
	if p.AppName != "" {
		c.AppName = p.AppName
	}

	if p.SourceRepoUrl != "" {
		c.SourceRepoUrl = p.SourceRepoUrl
	}

	if p.ProxyCollectionListPatch.Socks5 != nil {
		c.ProxyCollectionList.Socks5 = *p.ProxyCollectionListPatch.Socks5
	}

	if p.OptionsPatch.CacheDir != nil {
		c.Options.CacheDir = *p.OptionsPatch.CacheDir
	}
}
