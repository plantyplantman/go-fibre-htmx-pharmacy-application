package bigc

import (
	"bytes"
	"context"
	"errors"
	"text/template"

	"github.com/plantyplantman/bcapi/pkg/env"
	openai "github.com/sashabaranov/go-openai"
)

type AI_Client struct{ client *openai.Client }

const (
	KEYWORD_GENERATOR_PROMPT    = "You are a search engine optimisation specialist.\nGiven the name of a product, generate a list of search keywords that would a consumer may use to find the product.\nExample:\nProduct name: Ecoya Vanilla & Tonka Bean Madison Candle 400g\nKeywords: ecoya candle, ecoya, ecoya home fragrance, home fragrances, vanilla candle, vanilla bean, tonka bean, madison, candle, jar candle, spray, guava lychee, home spray, lotus flower"
	METADATA_GENERATOR_PROMPT   = "You are a search engine optimisation specialist.\nGiven the name of a product, generate a description of the product for the metadata of a webpage to assist in search engine optimisation.\nExample:\nProduct name: Ecoya Vanilla & Tonka Bean Madison Candle 400g\nMetadata description: Ecoya Vanilla & Tonka Bean Madison Candle's natural soy wax is blended with our signature fragrances and is poured into a contemporary and refined glass jar, providing a delicately scented burn time of up to 80 hours. With a decadent silver lid and presented in a beautifully designed box, the Madison Candle is the perfect gift for any occasion."
	NAME_SUGGESTOR_PROMPT       = "Given a truncated product name, suggest the full product name.\nExample:\nTruncated product name: E/PLAST SENS DRESS 3XL 10X15CM 5\nFull product name: Elastoplast Sensitive Dressing 3XL 10x15cm 5 Pack"
	CATEGORY_DESCRIPTION_PROMPT = `You are a search engine optimisation specialist.\nYou are working to write category descriptions for the online store of a pharmacy.\nGiven a list of categories and a target category, you will write html snippets which link the current category with the other categories in the list.\nThe descriptions you write for each category must:\n1. Contain common search key words relevant to the category\n2. Is highly relevant and descriptive\n3. Contain anchor tags linking to other categories provided.\nDO NOT exceed 180 words\nExample:<p>Treat your body like the temple it is with our variety of high quality products, designed to make you feel and look your best. Explore our exfoliators and <a href="https://www.thepharmacynetwork.com.au/bath-body/bath/body-scrubs/">body scrubs</a> to maintain rejuvenated skin, perfect for preparing your skin for <a href="https://www.thepharmacynetwork.com.au/bath-body/hair-removal/">hair removal</a> and <a href="https://www.thepharmacynetwork.com.au/bath-body/tanning/">tanning</a>.</p><p>Care for your nails like a professional with our home manicure sets and nail accessories. Discover the beauty tools to elevate your <a href="https://www.thepharmacynetwork.com.au/beauty-skin/hand-nail/">hand and nail care</a>. Begin a simple and effective skincare routine all year round with our <a href="https://www.thepharmacynetwork.com.au/beauty-skin/skincare/skincare-tools/">skincare tools</a>, available online and within our local pharmacies!</p>\n\nTARGET CATEGORY: {{with .TargetCategory}}\tName: {{.Name}}\n{{end}}\nRELATED CATEGORIES: {{with .RelatedCategories}}{{range .}}\n\tName: {{.Name}}\n\tUrl Snippet: {{with .CustomURL}}"{{.URL}}"\n{{end}}{{end}}{{end}}\nUse the following search keywords where relevant {{.Keywords}}`
)

type CategoryDescriptionPromptParams struct {
	TargetCategory    *Category
	RelatedCategories []*Category
	Keywords          []string
}

func GetOpenAiClient() (AI_Client, error) {
	if env.OPENAI == "" {
		return AI_Client{}, errors.New("FAILED TO READ OPENAI KEY FROM ENV")
	}
	c := openai.NewClient(env.OPENAI)
	return AI_Client{client: c}, nil
}

func (c AI_Client) getCompletion(messages []openai.ChatCompletionMessage, model string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c AI_Client) GenerateProductNameFromMinfosName(minfosName string, model string) (string, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	m := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: NAME_SUGGESTOR_PROMPT,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: minfosName,
		},
	}
	v, e := c.getCompletion(m, model)
	if e != nil {
		return "", e
	}

	v = RemoveSubstring(v, "Full product name: ")
	return v, nil
}

func (c AI_Client) GenerateMetadataFromProductName(productName string, model string) (string, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	m := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: METADATA_GENERATOR_PROMPT,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: productName,
		},
	}
	return c.getCompletion(m, model)
}

func (c AI_Client) GenerateSearchKeywordsFromProductName(productName string, model string) (string, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	m := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: KEYWORD_GENERATOR_PROMPT,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: productName,
		},
	}
	return c.getCompletion(m, model)
}

func (c AI_Client) GenerateCategoryDescriptionFromCategoryName(params CategoryDescriptionPromptParams, model string) (string, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	prompt, e := getCategoryDescriptionPrompt(params)
	if e != nil {
		return "", e
	}

	m := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "",
		},
	}
	return c.getCompletion(m, model)
}

func getCategoryDescriptionPrompt(params CategoryDescriptionPromptParams) (string, error) {
	t := template.Must(template.New("CategoryDescriptionPrompt").Parse(CATEGORY_DESCRIPTION_PROMPT))

	var tpl bytes.Buffer
	e := t.Execute(&tpl, params)
	if e != nil {
		return "", e
	}

	return tpl.String(), nil
}
