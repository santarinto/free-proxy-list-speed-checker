package commands

import (
	"fmt"
	"os"
	"reflect"

	"free-proxy-list-speed-checker/internal/config"
	"free-proxy-list-speed-checker/internal/network"
)

func Scan(cfg *config.Config) {
	collection := "socks5"
	if len(os.Args) > 2 {
		collection = os.Args[2]
	}

	if !collectionExists(collection, cfg) {
		fmt.Printf("Error: collection '%s' not found\n", collection)
		fmt.Println("\nAvailable collections:")
		List(cfg)
		os.Exit(1)
	}

	if err := network.Scan(collection, cfg); err != nil {
		fmt.Printf("Error during scan: %v\n", err)
		os.Exit(1)
	}
}

func collectionExists(collection string, cfg *config.Config) bool {
	v := reflect.ValueOf(cfg.ProxyCollectionList)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tagValue := field.Tag.Get("toml")
		if tagValue == collection {
			return true
		}
	}
	return false
}
