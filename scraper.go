package main

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
)

// це структура для постів Hacker News, Reddit, щоб можна було легше відображати дані
type Post struct {
	Title     string
	URL       string
	Points    string
	Author    string
	Comments  string
	TimeAgo   string
	Rank      int
	PostType  string
	Subreddit string 
	Site      string 
}

//кольори для відображення постів
var (
	titleColor     = color.New(color.FgHiCyan, color.Bold)
	urlColor       = color.New(color.FgBlue, color.Underline)
	rankColor      = color.New(color.FgYellow, color.Bold)
	authorColor    = color.New(color.FgGreen)
	pointsColor    = color.New(color.FgMagenta)
	commentsColor  = color.New(color.FgHiBlue)
	timeColor      = color.New(color.FgHiYellow)
	separatorColor = color.New(color.FgHiBlack)
	headerColor    = color.New(color.FgHiWhite, color.Bold)
	subredditColor = color.New(color.FgHiRed)
//кольори для різних типів постів 
	postTypeColors = map[string]*color.Color{
		"story":  color.New(color.FgWhite),
		"ask":    color.New(color.FgHiGreen),
		"show":   color.New(color.FgHiMagenta),
		"tell":   color.New(color.FgHiCyan),
		"launch": color.New(color.FgHiRed),
		"reddit": color.New(color.FgHiYellow),
	}
)

//scrapeHackerNews збирає пости з Hacker News.
func scrapeHackerNews(page int) {
	c := colly.NewCollector() 
	c.SetRequestTimeout(30 * time.Second) // таймаут на 30 секунд
	c.UserAgent = "Mozilla/5.0 (compatible; Go-Scraper/1.0)"  // пробую імітувати браузер щоб не блокнуло

	posts := make(map[string]*Post) //карта щоб зберігати пости

	c.OnHTML("tr.athing", func(e *colly.HTMLElement) { // збираємо дані з кожного поста
		id := e.Attr("id")
		title := e.ChildText("td.title > span.titleline > a")
		url := e.ChildAttr("td.title > span.titleline > a", "href")
		if strings.HasPrefix(url, "item?id=") { 
			url = "https://news.ycombinator.com/" + url // якщо посилання відносне, то додаємо URL
		}

		rankText := e.ChildText("span.rank") //тут ранг для hn і апвоути для реддіт
		rankNum := 0
		if rankText != "" {
			rankText = strings.TrimSuffix(rankText, ".")
			rankNum, _ = strconv.Atoi(rankText)
		}

		postType := "story" //який префікс такий і підтип
		switch {
		case strings.HasPrefix(title, "Ask HN:"):
			postType = "ask"
		case strings.HasPrefix(title, "Show HN:"):
			postType = "show"
		case strings.HasPrefix(title, "Tell HN:"):
			postType = "tell"
		case strings.HasPrefix(title, "Launch HN:"):
			postType = "launch"
		}

		posts[id] = &Post{ //зберігаємо уже в ту карту
			Title:    title,
			URL:      url,
			Rank:     rankNum,
			PostType: postType,
			Site:     "hn",
		}
	})

	c.OnHTML("tr.athing + tr", func(e *colly.HTMLElement) { //шукаємо бали,автор, час, коментарі
		id := e.DOM.Prev().AttrOr("id", "")
		if id == "" || posts[id] == nil { // я1кщо не знайшли відповідного id — пропускаємо
			return
		}
		post := posts[id]
		post.Points = e.ChildText("span.score")
		post.Author = e.ChildText("a.hnuser")

		post.TimeAgo = e.ChildText("span.age")
		post.Comments = ""
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			text := el.Text
			if strings.Contains(href, "item?id=") && strings.Contains(text, "comment") {
				post.Comments = text
			}
		})
	})


	// обробка помилок
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("HN, Не знайдено - URL: %s, Помилка: %v\n", r.Request.URL, err)
	})

	url := "https://news.ycombinator.com/"
	if page > 1 {
		url = fmt.Sprintf("https://news.ycombinator.com/news?p=%d", page)
	}
	printHeader("HACKER NEWS", fmt.Sprintf("СТОРІНКА %d", page))

	if err := c.Visit(url); err != nil {
		log.Fatal("Помилка:", err)
	}

	displayPosts(posts) // після завершення збору даних виводимо всі пости
}
//Тут майже все те саме
func scrapeReddit(subreddit string) {
	c := colly.NewCollector()
	c.SetRequestTimeout(30 * time.Second)
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)"

	var posts []*Post
	rank := 1

	c.OnHTML("div.thing", func(e *colly.HTMLElement) {
		title := e.ChildText("p.title > a.title")
		if title == "" {
			return
		}

		url := e.ChildAttr("p.title > a.title", "href")
		if strings.HasPrefix(url, "/r/") {
			url = "https://old.reddit.com" + url //old reddit тому що новий використовує реакт який просто поверне контейнер без списку

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
		log.Printf("Reddit не знайдено - URL: %s, Помилка: %v\n", r.Request.URL, err)
	})

	printHeader("REDDIT", fmt.Sprintf("r/%s", subreddit))
	url := fmt.Sprintf("https://old.reddit.com/r/%s", subreddit)
	if err := c.Visit(url); err != nil {
		log.Fatal("Помилка:", err)
	}

	
	postMap := make(map[string]*Post)
	for i, p := range posts {
		postMap[fmt.Sprintf("post_%d", i)] = p
	}
	displayPosts(postMap)
}


func printHeader(site, subtitle string) {
	headerColor.Printf("\n  ╔═══════════════════════════════════════════════════════╗\n") //не рухати бо не буде відображати рівно
	headerColor.Printf("  ║                    %-15s                    ║\n", site)		  //не рухати бо не буде відображати рівно
	headerColor.Printf("  ║                    %-15s                    ║\n", subtitle)   //не рухати бо не буде відображати рівно
	headerColor.Printf("  ╚═══════════════════════════════════════════════════════╝\n\n")
}


func displayPosts(posts map[string]*Post) {
	if len(posts) == 0 {
		color.Red("Не знайдено постів.\n")
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
