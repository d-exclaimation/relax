package quote

import (
	"encoding/json"
	"net/http"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
)

// Quote is the data structure for the quote API
type Quote struct {
	ID           string   `json:"_id"`
	Content      string   `json:"content"`
	Author       string   `json:"author"`
	Tags         []string `json:"tags"`
	AuthorSlug   string   `json:"authorSlug"`
	Length       int      `json:"length"`
	DateAdded    string   `json:"dateAdded"`
	DateModified string   `json:"dateModified"`
}

// Random is a function that returns a random quotes
func Random() async.Task[Quote] {
	return async.New(func() (Quote, error) {
		data := make([]Quote, 1)
		resp, err := http.Get(config.Env.QuoteAPI() + "/quotes/random?limit=1")
		if err != nil {
			return Quote{}, err
		}
		defer resp.Body.Close()

		json.NewDecoder(resp.Body).Decode(&data)
		return data[0], nil
	})
}
