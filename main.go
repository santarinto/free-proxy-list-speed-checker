package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"free-proxy-list-speed-checker/internal/cache"
	"free-proxy-list-speed-checker/internal/commands"
	"free-proxy-list-speed-checker/internal/config"
)

func printUsage() {
	fmt.Println("Free Proxy List Speed Checker")
	fmt.Println("\nUsage:")
	fmt.Println("  program <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  list")
	fmt.Println("      List all available proxy server collections")
	fmt.Println()
	fmt.Println("  scan <collection_name>")
	fmt.Println("      Scan a proxy server collection for speed testing")
	fmt.Println("      Arguments:")
	fmt.Println("        collection_name - Name of the collection (default: socks5)")
	fmt.Println()
	fmt.Println("  stats <collection_name>")
	fmt.Println("      Display available speed information for a collection")
	fmt.Println("      Arguments:")
	fmt.Println("        collection_name - Name of the collection (default: socks5)")
	fmt.Println()
	fmt.Println("  get-fast <collection_name> <number>")
	fmt.Println("      Get the fastest proxy servers from a collection")
	fmt.Println("      Arguments:")
	fmt.Println("        collection_name - Name of the collection (default: socks5)")
	fmt.Println("        number          - Number of proxies to retrieve (default: 1)")
	fmt.Println()
	fmt.Println("  clear")
	fmt.Println("      Clear the cache")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  program list")
	fmt.Println("  program scan socks5")
	fmt.Println("  program stats")
	fmt.Println("  program get-fast socks5 5")
	fmt.Println("  program clear")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cache.New(cfg.Options.CacheDir)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("cache close failed: %v", err)
		}
	}()

	switch command {
	case "list":
		commands.List(cfg)

	case "scan":
		commands.Scan(cfg)

	case "stats":
		collection := "socks5"
		if len(os.Args) > 2 {
			collection = os.Args[2]
		}
		fmt.Printf("Displaying stats for collection: %s\n", collection)

	case "get-fast":
		collection := "socks5"
		number := 1
		if len(os.Args) > 2 {
			collection = os.Args[2]
		}
		if len(os.Args) > 3 {
			if n, err := strconv.Atoi(os.Args[3]); err == nil {
				number = n
			}
		}
		fmt.Printf("Getting %d fastest proxy(s) from collection: %s\n", number, collection)

	case "clear":
		fmt.Println("Clearing cache...")

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Printf("\nCache directory: %s\n", c.Dir)
		printUsage()
		os.Exit(1)
	}
}
