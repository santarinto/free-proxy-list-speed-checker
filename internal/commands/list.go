package commands

import (
	"fmt"
	"reflect"
	"strings"

	"free-proxy-list-speed-checker/internal/config"
)

func List(cfg *config.Config) {
	fmt.Println("Available proxy collections:")
	v := reflect.ValueOf(cfg.ProxyCollectionList)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tagValue := field.Tag.Get("toml")
		if tagValue != "" {
			fmt.Printf("  - %s\n", tagValue)
		} else {
			fmt.Printf("  - %s\n", strings.ToLower(field.Name))
		}
	}
}
