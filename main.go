package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type Page struct {
	Title string `json:"title"`
}

type QueryResult struct {
	Pages []Page `json:"allpages"`
}

type APIResponse struct {
	Query QueryResult `json:"query"`
}

func fetchAllPages(apiURL string) ([]string, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch pages: %s", resp.Status)
	}

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return nil, err
	}

	var pages []string
	for _, page := range apiResp.Query.Pages {
		pages = append(pages, page.Title)
	}

	return pages, nil
}

func fetchWikiPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page: %s", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return text.String(), nil
}

func saveToFile(text, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}

func main() {
	apiURL := "https://nikke-goddess-of-victory-international.fandom.com/api.php?action=query&list=allpages&aplimit=max&format=json"
	pages, err := fetchAllPages(apiURL)
	if err != nil {
		fmt.Println("Error fetching all pages:", err)
		return
	}

	for _, title := range pages {
		url := fmt.Sprintf("https://nikke-goddess-of-victory-international.fandom.com/wiki/%s", strings.ReplaceAll(title, " ", "_")) // Wiki page
		text, err := fetchWikiPage(url)
		if err != nil {
			fmt.Println("Error fetching wiki page:", err)
			continue
		}

		filename := fmt.Sprintf("%s.txt", strings.ReplaceAll(title, "/", "_")) // File name for Windows, since / does not work, I replaced it with _
		err = saveToFile(text, filename)
		if err != nil {
			fmt.Println("Error saving to file:", err)
		}
	}
}
