package scraping

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
	Tier    int        `json:"tier"`
}

func ScrapeDatafromWeb() ([]Element, error) {
	res, err := http.Get("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var elements []Element
	tier := 0

	doc.Find(".mw-parser-output").Children().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "h3" {
			headerText := strings.TrimSpace(s.Text())
			if strings.HasPrefix(headerText, "Tier ") {
				parts := strings.Split(headerText, " ")
				if len(parts) >= 2 {
					tier, _ = strconv.Atoi(parts[1])
				}
			}
		} else if goquery.NodeName(s) == "table" {
			s.Find("tr").Each(func(i int, tr *goquery.Selection) {
				tds := tr.Find("td")
				if tds.Length() >= 2 {
					name := strings.TrimSpace(tds.Eq(0).Find("a").Text())
					if name == "" {
						return
					}
					var recipes [][]string
					tds.Eq(1).Find("li").Each(func(i int, li *goquery.Selection) {
						var pair []string
						li.Find("a").Each(func(j int, a *goquery.Selection) {
							ingredient := strings.TrimSpace(a.Text())
							if ingredient != "" {
								pair = append(pair, ingredient)
							}
						})
						if len(pair) == 2 {
							recipes = append(recipes, pair)
						}
					})
					elements = append(elements, Element{
						Name:    name,
						Recipes: recipes,
						Tier:    tier,
					})
				}
			})
		}
	})

	elements = filterElements(elements)
	elements = removeInvalidRecipes(elements)
	elements = filterElements(elements)
	elements = removeInvalidRecipes(elements)
	elements = filterElements(elements)

	return elements, nil
}

func filterElements(elements []Element) []Element {
	var cleaned []Element
	for _, el := range elements {
		if el.Name == "Little Alchemy 1" {
			continue
		}
		if len(el.Recipes) > 0 || el.Name == "Air" || el.Name == "Earth" || el.Name == "Fire" || el.Name == "Water" {
			cleaned = append(cleaned, el)
		}
	}
	return cleaned
}

func removeInvalidRecipes(elements []Element) []Element {
	validNames := make(map[string]bool)
	for _, el := range elements {
		validNames[el.Name] = true
	}

	for i := range elements {
		var validRecipes [][]string
		for _, recipe := range elements[i].Recipes {
			if len(recipe) == 2 && validNames[recipe[0]] && validNames[recipe[1]] {
				validRecipes = append(validRecipes, recipe)
			}
		}
		elements[i].Recipes = validRecipes
	}

	return elements
}
