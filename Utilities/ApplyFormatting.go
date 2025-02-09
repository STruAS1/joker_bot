package Utilities

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RemoveHTMLTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

var fancyDigits = map[rune]rune{
	'0': '𝟘', '1': '𝟙', '2': '𝟚', '3': '𝟛', '4': '𝟜',
	'5': '𝟝', '6': '𝟞', '7': '𝟟', '8': '𝟠', '9': '𝟡',
	'.': '.',
}

func ConvertToFancyString(num int) string {
	numStr := fmt.Sprintf("%d", num)
	var builder strings.Builder
	for _, digit := range numStr {
		if fancy, ok := fancyDigits[digit]; ok {
			builder.WriteRune(fancy)
		} else {
			builder.WriteRune(digit)
		}
	}
	return builder.String()
}
func ConvertToFancyStringFloat(numStr string) string {
	numStr = strings.TrimRight(numStr, "0")
	numStr = strings.TrimRight(numStr, ".")

	var builder strings.Builder
	for _, digit := range numStr {
		if fancy, ok := fancyDigits[digit]; ok {
			builder.WriteRune(fancy)
		} else {
			builder.WriteRune(digit)
		}
	}
	return builder.String()
}
func ApplyFormatting(text string, entities []tgbotapi.MessageEntity) string {
	text = RemoveHTMLTags(text)

	var formattedText strings.Builder
	entityMap := make(map[int][]tgbotapi.MessageEntity)

	runes := []rune(text)

	for _, entity := range entities {
		for i := entity.Offset; i < entity.Offset+entity.Length; i++ {
			entityMap[i] = append(entityMap[i], entity)
		}
	}

	i := 0
	for i < len(runes) {
		if entityGroup, exists := entityMap[i]; exists {
			var entityText strings.Builder

			for i < len(runes) && entityMap[i] != nil {
				entityText.WriteRune(runes[i])
				i++
			}

			result := entityText.String()

			sort.SliceStable(entityGroup, func(a, b int) bool {
				return entityGroup[a].Offset < entityGroup[b].Offset
			})

			for _, entity := range entityGroup {
				switch entity.Type {
				case "bold":
					result = fmt.Sprintf("<b>%s</b>", result)
				case "italic":
					result = fmt.Sprintf("<i>%s</i>", result)
				case "underline":
					result = fmt.Sprintf("<u>%s</u>", result)
				case "strikethrough":
					result = fmt.Sprintf("<s>%s</s>", result)
				case "spoiler":
					result = fmt.Sprintf("<span class=\"tg-spoiler\">%s</span>", result)
				case "blockquote":
					result = fmt.Sprintf("<blockquote>%s</blockquote>", result)
				case "expandable_blockquote":
					result = fmt.Sprintf("<blockquote expandable>%s</blockquote>", result)
				case "code":
					result = fmt.Sprintf("<code>%s</code>", result)
				case "pre":
					result = fmt.Sprintf("<pre>%s</pre>", result)
				case "text_link":
					result = fmt.Sprintf("<a href='%s'>%s</a>", entity.URL, result)
				case "text_mention":
					if entity.User != nil {
						result = fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", entity.User.ID, result)
					}
				case "url":
					result = fmt.Sprintf("<a href='%s'>%s</a>", result, result)
				case "email":
					result = fmt.Sprintf("<a href='mailto:%s'>%s</a>", result, result)
				case "phone_number":
					result = fmt.Sprintf("<a href='tel:%s'>%s</a>", result, result)
				}
			}

			formattedText.WriteString(result)
		} else {
			formattedText.WriteRune(runes[i])
			i++
		}
	}

	return formattedText.String()
}
