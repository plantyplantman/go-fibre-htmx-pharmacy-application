package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/sashabaranov/go-openai"
)

func main() {
	f, err := os.Open("cmd/bigc/seo/vitaminKwargs.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	var keywords = make([]string, 0)
	for _, line := range lines {
		keywords = append(keywords, strings.TrimSpace(line[0]))
	}

	c := bigc.MustGetClient()
	root, err := c.GetCategoryFromID(32)
	if err != nil {
		log.Fatalln(err)
	}

	kids, err := root.GetChildren(c)
	if err != nil {
		log.Fatalln(err)
	}

	ai_c, err := bigc.GetOpenAiClient()
	if err != nil {
		log.Fatalln(err)
	}

	for _, kid := range kids {
		if kid.Description == "" {
			descCh := make(chan string, 1)
			go func(k *bigc.Category) {
				desc, err := ai_c.GenerateCategoryDescriptionFromCategoryName(bigc.CategoryDescriptionPromptParams{
					TargetCategory:    &kid,
					RelatedCategories: kids,
					Keywords:          keywords,
				}, openai.GPT4)
				if err != nil {
					log.Println(err)
				}
				descCh <- desc
			}(&kid)

			kwargCh := make(chan string, 1)
			go func(name string) {
				kw, err := ai_c.GenerateCategoryKeywordsFromCategoryName(name, openai.GPT4)
				if err != nil {
					log.Println(err)
				}
				kwargCh <- kw
			}(kid.Name)

			metaCh := make(chan string, 1)
			go func(name string) {
				meta, err := ai_c.GenerateCategoryMetaFromCategoryName(name, openai.GPT4)
				if err != nil {
					log.Println(err)
				}
				metaCh <- meta
			}(kid.Name)

			_, err = c.UpdateCategory(&kid, bigc.WithUpdateDesc(<-descCh), bigc.WithUpdateMetaDesc(<-metaCh), bigc.WithUpdateMetaKeywords(strings.Split(<-kwargCh, ",")))
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
