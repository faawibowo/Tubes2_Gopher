package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

/* ──────────── Public Types ──────────── */

type ElementJSON struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

type ScrapedData struct {
	Ingredients []string
	Tier        int
}

/* ──────────── Load JSON ──────────── */

func LoadElementsJSON(path string) ([]ElementJSON, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading json: %w", err)
	}

	var result []ElementJSON
	if err := json.Unmarshal(buf, &result); err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}
	return result, nil
}

/* ──────────── Scraping + Tier helper ──────────── */

func ToScrapedMapWithTiers(elems []ElementJSON) map[string]ScrapedData {
	// convert to the lightweight struct used by ComputeTiers
	simple := make([]Element, len(elems))
	for i, e := range elems {
		simple[i] = Element{Name: e.Name, Recipes: e.Recipes}
	}

	tierMap, _ := ComputeTiers(simple)

	out := make(map[string]ScrapedData)
	for _, e := range elems {
		var ing []string
		for _, pair := range e.Recipes {
			if len(pair) == 2 {
				ing = append(ing, pair[0], pair[1])
			}
		}
		out[e.Name] = ScrapedData{
			Ingredients: ing,
			Tier:        tierMap[e.Name], // if absent ⇒ 0
		}
	}
	return out
}

/* ──────────── Internal helper types ──────────── */

type Element struct {
	Name    string
	Recipes [][]string
}

/* ──────────── Tier computation ──────────── */

// ComputeTiers assigns:
//
//	Tier-0  →  "Air", "Earth", "Water", "Fire"
//	Tier-n  →  smallest n such that a recipe exists whose every ingredient tier < n
//
// Returns: (tier map, list of unreachable / cyclic names)
func ComputeTiers(elems []Element) (map[string]int, []string) {
	// 1. prepare recipe lookup
	recipes := make(map[string][][]string, len(elems))
	for _, el := range elems {
		recipes[el.Name] = el.Recipes
	}

	// 2. Tier-0 seeds
	basics := map[string]struct{}{
		"Air": {}, "Earth": {}, "Water": {}, "Fire": {},
	}

	tier := make(map[string]int, len(elems))
	remaining := make(map[string]struct{}, len(elems))

	for _, el := range elems {
		if _, ok := basics[el.Name]; ok {
			tier[el.Name] = 0
		} else {
			remaining[el.Name] = struct{}{}
		}
	}

	// 3. iterative relaxation (BFS on tiers)
	for changed := true; changed; {
		changed = false
		for name := range remaining {
			for _, r := range recipes[name] {
				ok, maxDep := true, -1
				for _, ing := range r {
					t, has := tier[ing]
					if !has {
						ok = false
						break
					}
					if t > maxDep {
						maxDep = t
					}
				}
				if ok { // all ingredients already have tiers
					tier[name] = maxDep + 1
					delete(remaining, name)
					changed = true
					break
				}
			}
		}
	}

	// 4. whatever’s left is cyclic / unreachable
	cycles := make([]string, 0, len(remaining))
	for n := range remaining {
		cycles = append(cycles, n)
	}
	sort.Strings(cycles)
	return tier, cycles
}
