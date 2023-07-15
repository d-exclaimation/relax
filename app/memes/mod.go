package memes

import (
	"encoding/json"
	"net/http"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
)

// Meme is the data structure for the memes API
type Meme struct {
	PostLink  string   `json:"postLink"`
	Subreddit string   `json:"subreddit"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Nsfw      bool     `json:"nsfw"`
	Spoiler   bool     `json:"spoiler"`
	Author    string   `json:"author"`
	Ups       int      `json:"ups"`
	Preview   []string `json:"preview"`
}

// Random is a function that returns a random quotes
func Random() async.Task[Meme] {
	return async.New(func() (Meme, error) {
		data := Meme{}
		for {
			resp, err := http.Get(config.Env.MemeAPI())
			if err != nil {
				return data, err
			}
			defer resp.Body.Close()

			json.NewDecoder(resp.Body).Decode(&data)

			if !data.Nsfw {
				return data, nil
			}
		}
	})
}
