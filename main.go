package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	site := strings.ToLower(os.Args[1])
	page := 1

	switch site {
	case "hn", "hackernews":
		if len(os.Args) > 2 {
			if pageArg, err := strconv.Atoi(os.Args[2]); err == nil && pageArg > 0 {
				page = pageArg
			}
		}
		scrapeHackerNews(page)

	case "reddit", "r":
		subreddit := "popular"
		if len(os.Args) > 2 {
			subreddit = os.Args[2]
		}
		scrapeReddit(subreddit)

	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════════════════════╗")
	fmt.Println("  ║                   МУЛЬТИ-САЙТ СКРАППЕР                ║")
	fmt.Println("  ╚═══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Використання:")
	fmt.Println("  ./scraper hn [page]              # Hacker News (Дефолт: Сторінка 1)")
	fmt.Println("  ./scraper reddit [subreddit]     # Reddit (Дефолт: popular)")
	fmt.Println()
	fmt.Println("Приклад:")
	fmt.Println("  ./scraper hn 2                   # Hacker News сторінка 2")
	fmt.Println("  ./scraper reddit golang          # r/golang сабреддіт")
	fmt.Println("  ./scraper reddit programming     # r/programming сабреддіт")
	fmt.Println()
}
