package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
)

type Post struct {
	Title     string
	URL       string
	Points    string
	Author    string
	Comments  string
	TimeAgo   string
	Rank      int
	PostType  string
	Subreddit string // For Reddit posts
	Site      string // "hn" or "reddit"
}

var (
	titleColor      = color.New(color.FgHiCyan, color.Bold)
	urlColor        = color.New(color.FgBlue, color.Underline)
	rankColor       = color.New(color.FgYellow, color.Bold)
	authorColor     = color.New(color.FgGreen)
	pointsColor     = color.New(color.FgMagenta)
	commentsColor   = color.New(color.FgHiBlue)
	timeColor       = color.New(color.FgHiYellow)
	separatorColor  = color.New(color.FgHiBlack)
	headerColor     = color.New(color.FgHiWhite, color.Bold)
	subredditColor  = color.New(color.FgHiRed)
	postTypeColors = map[string]*color.Color{
		"story":  color.New(color.FgWhite),
		"ask":    color.New(color.FgHiGreen),
		"show":   color.New(color.FgHiMagenta),
		"tell":   color.New(color.FgHiCyan),
		"launch": color.New(color.FgHiRed),
		"reddit": color.New(color.FgHiYellow),
	}
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
	headerColor.Println("\n╔═══════════════════════════════════════════════════════╗")
	headerColor.Println("  ║                   MULTI-SITE SCRAPER                 ║")
	headerColor.Println("  ╚═══════════════════════════════════════════════════════╝")
	fmt.Println()
	headerColor.Println("Usage:")
	fmt.Println("  ./scraper hn [page]              # Hacker News (default: page 1)")
	fmt.Println("  ./scraper reddit [subreddit]     # Reddit (default: popular)")
	fmt.Println()
	headerColor.Println("Examples:")
	fmt.Println("  ./scraper hn 2                   # Hacker News page 2")
	fmt.Println("  ./scraper reddit golang          # r/golang subreddit")
	fmt.Println("  ./scraper reddit programming     # r/programming subreddit")
	fmt.Println()
}

func scrapeHackerNews(page int) {
	c := colly.NewCollector()
	c.SetRequestTimeout(30 * time.Second)
	posts := make(map[string]*Post)

	// Set User-Agent
	c.UserAgent = "Mozilla/5.0 (compatible; Go-Scraper/1.0)"

	c.OnHTML("tr.athing", func(e *colly.HTMLElement) {
		id := e.Attr("id")
		title := e.ChildText("td.title > span.titleline > a")
		url := e.ChildAttr("td.title > span.titleline > a", "href")

		// Handle relative URLs
		if strings.HasPrefix(url, "item?id=") {
			url = "https://news.ycombinator.com/" + url
		}

		rankText := e.ChildText("span.rank")
		rankNum := 0
		if rankText != "" {
			rankText = strings.TrimSuffix(rankText, ".")
			rankNum, _ = strconv.Atoi(rankText)
		}

		postType := "story"
		if strings.HasPrefix(title, "Ask HN:") {
			postType = "ask"
		} else if strings.HasPrefix(title, "Show HN:") {
			postType = "show"
		} else if strings.HasPrefix(title, "Tell HN:") {
			postType = "tell"
		} else if strings.HasPrefix(title, "Launch HN:") {
			postType = "launch"
		}

		posts[id] = &Post{
			Title:    title,
			URL:      url,
			Rank:     rankNum,
			PostType: postType,
			Site:     "hn",
		}
	})

	c.OnHTML("tr.athing + tr", func(e *colly.HTMLElement) {
		id := e.DOM.Prev().AttrOr("id", "")
		if id == "" || posts[id] == nil {
			return
		}

		post := posts[id]
		post.Points = e.ChildText("span.score")
		post.Author = e.ChildText("a.hnuser")

		e.ForEach("span.age", func(_ int, el *colly.HTMLElement) {
			post.TimeAgo = el.Text
		})

		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			text := el.Text
			if strings.Contains(href, "item?id=") && strings.Contains(text, "comment") {
				post.Comments = text
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("HN Error - URL: %s, Error: %v\n", r.Request.URL, err)
	})

	url := "https://news.ycombinator.com/"
	if page > 1 {
		url = fmt.Sprintf("https://news.ycombinator.com/news?p=%d", page)
	}

	printHeader("HACKER NEWS", fmt.Sprintf("PAGE %d", page))

	err := c.Visit(url)
	if err != nil {
		log.Fatal("Failed to scrape Hacker News:", err)
	}

	displayPosts(posts)
}

func scrapeReddit(subreddit string) {
	c := colly.NewCollector()
	c.SetRequestTimeout(30 * time.Second)
	
	// Important: Reddit requires proper User-Agent
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	
	var posts []*Post
	rank := 1

	// Reddit's old interface is more scraper-friendly
	c.OnHTML("div.thing", func(e *colly.HTMLElement) {
		title := e.ChildText("p.title > a.title")
		if title == "" {
			return
		}

		url := e.ChildAttr("p.title > a.title", "href")
		
		// Handle Reddit internal links
		if strings.HasPrefix(url, "/r/") {
			url = "https://old.reddit.com" + url
		}

		author := e.ChildText("p.tagline > a.author")
		points := e.ChildText("div.score.unvoted")
		if points == "•" {
			points = e.ChildText("div.score")
		}
		
		timeAgo := e.ChildAttr("p.tagline > time", "title")
		if timeAgo == "" {
			timeAgo = e.ChildText("p.tagline > time")
		}

		comments := e.ChildText("ul.flat-list > li > a[data-event-action='comments']")

		post := &Post{
			Title:     title,
			URL:       url,
			Points:    points,
			Author:    author,
			Comments:  comments,
			TimeAgo:   timeAgo,
			Rank:      rank,
			PostType:  "reddit",
			Subreddit: subreddit,
			Site:      "reddit",
		}

		posts = append(posts, post)
		rank++
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Reddit Error - URL: %s, Error: %v\n", r.Request.URL, err)
	})

	url := fmt.Sprintf("https://old.reddit.com/r/%s", subreddit)
	
	printHeader("REDDIT", fmt.Sprintf("r/%s", subreddit))

	err := c.Visit(url)
	if err != nil {
		log.Fatal("Failed to scrape Reddit:", err)
	}

	// Convert slice to map for consistent display
	postMap := make(map[string]*Post)
	for i, post := range posts {
		postMap[fmt.Sprintf("post_%d", i)] = post
	}

	displayPosts(postMap)
}

func printHeader(site, subtitle string) {
	headerColor.Printf("\n╔═══════════════════════════════════════════════════════╗\n")
	headerColor.Printf("  ║                    %-15s                    ║\n", site)
	headerColor.Printf("  ║                    %-15s                    ║\n", subtitle)
	headerColor.Printf("  ╚═══════════════════════════════════════════════════════╝\n\n")
}

func displayPosts(posts map[string]*Post) {
	if len(posts) == 0 {
		color.Red("No posts found. The site might be blocking requests or the structure changed.\n")
		return
	}

	for _, post := range posts {
		fmt.Println()
		rankColor.Printf(" #%-2d ", post.Rank)
		
		postTypeColor := postTypeColors[post.PostType]
		if post.Site == "reddit" {
			postTypeColor.Printf("[REDDIT] ")
			if post.Subreddit != "" {
				subredditColor.Printf("r/%s ", post.Subreddit)
			}
		} else {
			postTypeColor.Printf("[%s] ", strings.ToUpper(post.PostType))
		}
		
		titleColor.Println(post.Title)

		fmt.Print("     ")
		urlColor.Println(post.URL)

		fmt.Print("     ")
		if post.Points != "" && post.Points != "•" {
			pointsColor.Printf("%s • ", post.Points)
		}

		if post.Author != "" {
			authorColor.Printf("by %s • ", post.Author)
		}

		if post.TimeAgo != "" {
			timeColor.Printf("%s", post.TimeAgo)
		}

		if post.Comments != "" {
			fmt.Print(" • ")
			commentsColor.Print(post.Comments)
		}

		fmt.Println()
		separatorColor.Println(strings.Repeat("─", 75))
	}

	fmt.Println()
}