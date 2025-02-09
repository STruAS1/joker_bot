package Utilities

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func extractFirstTitle(htmlData string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlData))
	if err != nil {
		return "", err
	}

	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if title != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	if title == "" {
		return "", fmt.Errorf("тег <title> не найден")
	}
	return title, nil
}

func IsUsernameValid(username string) (bool, error) {
	username = strings.TrimPrefix(username, "@")
	url := fmt.Sprintf("https://t.me/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	body := string(bodyBytes)

	title, err := extractFirstTitle(body)
	if err != nil {
		return false, err
	}

	if strings.Contains(title, "Telegram Messenger") {
		return false, nil
	}
	return true, nil
}
