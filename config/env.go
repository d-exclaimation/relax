package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	OAUTH          = "OAUTH_TOKEN"
	OAUTH_APP      = "OAUTH_APP_TOKEN"
	OAUTH_APP_NAME = "OAUTH_APP_NAME"
	QUOTE_API_URL  = "QUOTE_API_URL"
	CHANNELS       = "CHANNEL_IDS"
	GO_ENV         = "GO_ENV"
)

// Environment is a struct that holds the environment variables
type Environment struct {
	oauth    string
	channels []string
	mode     string
	oauthApp string
	quoteAPI string
}

// Env is a global environment variables
var Env = &Environment{}

// Load loads the environment variables and stores them in the Env struct
func (*Environment) Load() {
	mode := GetGoEnv()
	if mode == "development" {
		godotenv.Load(".env")
	}
	Env.oauth = GetOAuthEnv()
	Env.mode = mode
	Env.channels = GetChannelsEnv()
	Env.oauthApp = GetOAuthAppEnv()
	Env.quoteAPI = GetQuoteAPIURL()
}

// OAuth lazily load and returns the OAuth token
func (e *Environment) OAuth() string {
	res := e.oauth
	if res == "" {
		res = GetOAuthEnv()
	}
	return res
}

// OAuthApp lazily load and returns the OAuth app token
func (e *Environment) OAuthApp() string {
	res := e.oauthApp
	if res == "" {
		res = GetOAuthAppEnv()
	}
	return res
}

// Channels lazily load andreturns the channels
func (e *Environment) Channels() []string {
	res := e.channels
	if res == nil {
		res = GetChannelsEnv()
	}
	return res
}

// Mode lazily load and returns the mode
func (e *Environment) Mode() string {
	res := e.mode
	if res == "" {
		res = GetGoEnv()
	}
	return res
}

// QuoteAPI lazily load and returns the quote API URL
func (e *Environment) QuoteAPI() string {
	res := e.quoteAPI
	if res == "" {
		res = GetQuoteAPIURL()
	}
	return res
}

// IsProduction returns true if the mode is production
func (e *Environment) IsProduction() bool {
	return e.Mode() == "production"
}

// GetGoEnv returns the mode from the environment directly
func GetGoEnv() string {
	res := os.Getenv(GO_ENV)
	if res == "" {
		res = "development"
	}
	return res
}

// GetOAuthEnv returns the OAuth token from the environment directly
func GetOAuthEnv() string {
	return os.Getenv(OAUTH)
}

// GetChannelsEnv returns the channels from the environment directly
func GetChannelsEnv() []string {
	res := os.Getenv(CHANNELS)
	return strings.Split(res, ",")
}

// GetOAuthAppEnv returns the OAuth app token from the environment directly
func GetOAuthAppEnv() string {
	return os.Getenv(OAUTH_APP)
}

// GetOAuthAppNameEnv returns the OAuth app name from the environment directly
func GetOAuthAppNameEnv() string {
	return os.Getenv(OAUTH_APP_NAME)
}

// GetQuoteAPIURL returns the quote API URL from the environment directly
func GetQuoteAPIURL() string {
	return os.Getenv(QUOTE_API_URL)
}
