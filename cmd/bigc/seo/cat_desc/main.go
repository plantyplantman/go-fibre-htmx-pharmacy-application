package main

import (
	"log"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/sashabaranov/go-openai"
)

func main() {
	var keywords = []string{
		"instant ice packs",
		"ice packs first aid",
		"first aid kit bag",
		"first aid bag",
		"medical kit",
		"sports first aid kit",
		"first aid for sport",
		"basic first aid kit",
		"first aid items",
		"first aid kit box",
		"burn kit",
		"emergency first aid",
		"first aid box",
		"small first aid kit",
		"burns first aid kit",
		"trauma kit",
		"first aid kit",
		"1st aid kit",
		"first aid pack",
		"first aid",
	}

	c := bigc.MustGetClient()
	root, err := c.GetCategoryFromID(58)
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
			desc, err := ai_c.GenerateCategoryDescriptionFromCategoryName(bigc.CategoryDescriptionPromptParams{
				TargetCategory:    &kid,
				RelatedCategories: kids,
				Keywords:          keywords,
			}, openai.GPT4)
			if err != nil {
				log.Println(err)
				continue
			}

			kw, err := ai_c.GenerateCategoryKeywordsFromCategoryName(kid.Name, openai.GPT4)
			if err != nil {
				log.Println(err)
				continue
			}

			meta, err := ai_c.GenerateCategoryMetaFromCategoryName(kid.Name, openai.GPT4)
			if err != nil {
				log.Println(err)
				continue
			}

			_, err = c.UpdateCategory(&kid, bigc.WithUpdateDesc(desc), bigc.WithUpdateMetaDesc(meta), bigc.WithUpdateMetaKeywords(strings.Split(kw, ",")))
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

// func genCats(root *bigc.Category, c *bigc.BigCommerceClient, ai_c *bigc.AI_Client, keywords []string) error {
// 	kids, err := root.GetChildren(c)
// 	if err != nil {
// 		return err
// 	}

// 	wg := &sync.WaitGroup{}
// 	errch := make(chan error)
// 	defer close(errch)
// 	for _, kid := range kids {
// 		if kid.Description == "" {
// 			wg.Add(1)

// 			go func(kid bigc.Category) {
// 				desc, err := ai_c.GenerateCategoryDescriptionFromCategoryName(bigc.CategoryDescriptionPromptParams{
// 					TargetCategory:    &kid,
// 					RelatedCategories: kids,
// 					Keywords:          keywords,
// 				}, openai.GPT4)
// 				if err != nil {
// 					errch <- err
// 				}

// 				kw, err := ai_c.GenerateCategoryKeywordsFromCategoryName(kid.Name, openai.GPT4)
// 				if err != nil {
// 					errch <- err
// 				}

// 				meta, err := ai_c.GenerateCategoryMetaFromCategoryName(kid.Name, openai.GPT4)
// 				if err != nil {
// 					errch <- err
// 				}

// 				_, err = c.UpdateCategory(&kid, bigc.WithUpdateDesc(desc), bigc.WithUpdateMetaDesc(meta), bigc.WithUpdateMetaKeywords(strings.Split(kw, ",")))
// 				if err != nil {
// 					errch <- err
// 				}

// 				wg.Done()
// 			}(kid)
// 		}
// 	}

// 	go func() {
// 		wg.Wait()
// 	}()

// 	select {
// 	case err := <-errch:
// 		return err
// 	default:
// 	}

// 	return nil
// }
