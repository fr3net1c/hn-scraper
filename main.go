package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
	"strconv"
	"strings"
)

type Story struct {
	Title     string
	URL       string
	Points    string
	Author    string
	Comments  string
	TimeAgo   string
	Rank      int
	StoryType string
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
	storyTypeColors = map[string]*color.Color{
		"story":  color.New(color.FgWhite),
		"ask":    color.New(color.FgHiGreen),
		"show":   color.New(color.FgHiMagenta),
		"tell":   color.New(color.FgHiCyan),
		"launch": color.New(color.FgHiRed),
	}
)

func main() {
	page := 1
	if len(os.Args) > 1 {
		pageArg, err := strconv.Atoi(os.Args[1])
		if err == nil && pageArg > 0 {
			page = pageArg
		}
	}

	c := colly.NewCollector()
	stories := make(map[string]*Story)
	
	c.OnHTML("tr.athing", func(e *colly.HTMLElement) {
		id := e.Attr("id")
		title := e.ChildText("td.title > span.titleline > a")
		url := e.ChildAttr("td.title > span.titleline > a", "href")
		
		rankText := e.ChildText("span.rank")
		rankNum := 0
		if rankText != "" {
			rankText = strings.TrimSuffix(rankText, ".")
			rankNum, _ = strconv.Atoi(rankText)
		}
		
		storyType := "story"
		if strings.HasPrefix(title, "Ask HN:") {
			storyType = "ask"
		} else if strings.HasPrefix(title, "Show HN:") {
			storyType = "show"
		} else if strings.HasPrefix(title, "Tell HN:") {
			storyType = "tell"
		} else if strings.HasPrefix(title, "Launch HN:") {
			storyType = "launch"
		}
		
		stories[id] = &Story{
			Title:     title,
			URL:       url,
			Rank:      rankNum,
			StoryType: storyType,
		}
	})
	
	c.OnHTML("tr.athing + tr", func(e *colly.HTMLElement) {
		id := e.DOM.Prev().AttrOr("id", "")
		if id == "" || stories[id] == nil {
			return
		}
		
		story := stories[id]
		story.Points = e.ChildText("span.score")
		story.Author = e.ChildText("a.hnuser")
		
		e.ForEach("span.age", func(_ int, el *colly.HTMLElement) {
			story.TimeAgo = el.Text
		})
		
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			text := el.Text
			if strings.Contains(href, "item?id=") && strings.Contains(text, "comment") {
				story.Comments = text
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("URL:", r.Request.URL, "Поломалось:", r, "\nОшибка:", err)
	})

	url := "https://news.ycombinator.com/"
	if page > 1 {
		url = fmt.Sprintf("https://news.ycombinator.com/news?p=%d", page)
	}
	
	headerColor.Printf("\n╔═══════════════════════════════════════════════════════╗\n")
	headerColor.Printf("  ║                       PAGE %-3d                       ║\n", page)
	headerColor.Printf("  ╚═══════════════════════════════════════════════════════╝\n\n")
	
	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, story := range stories {
		fmt.Println()
		rankColor.Printf(" #%-2d ", story.Rank)
		storyTypeColor := storyTypeColors[story.StoryType]
		storyTypeColor.Printf("[%s] ", strings.ToUpper(story.StoryType))
		titleColor.Println(story.Title)
		
		fmt.Print("     ")
		urlColor.Println(story.URL)
		
		fmt.Print("     ")
		if story.Points != "" {
			pointsColor.Printf("%s • ", story.Points)
		}
		
		if story.Author != "" {
			authorColor.Printf("by %s • ", story.Author)
		}
		
		if story.TimeAgo != "" {
			timeColor.Printf("%s", story.TimeAgo)
		}
		
		if story.Comments != "" {
			fmt.Print(" • ")
			commentsColor.Print(story.Comments)
		}
		
		fmt.Println()
		separatorColor.Println(strings.Repeat("─", 75))
	}
	
	fmt.Println()
	headerColor.Printf("Как использовать:  ./[название бинарника] [страница]\n")
	headerColor.Printf("Example: ./hn 2    # Страница 2\n\n")
}
