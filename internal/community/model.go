// Package community manages community participation profiles, validation, local storage,
// machine UUID derivation, and synchronization with the community API.
package community

import (
	"os"
	"strings"
)

const (
	MaxBioLength              = 200
	DefaultCommunityEndpoint  = "https://api.penguins-eggs.net/api/community"
	FallbackCommunityEndpoint = "https://penguins-eggs.net/api/community"
)

// CommunitySocials holds individual social platform handles in alphabetical order.
type CommunitySocials struct {
	Facebook string `yaml:"facebook,omitempty" json:"facebook,omitempty"`
	Gitee    string `yaml:"gitee,omitempty" json:"gitee,omitempty"`
	GitHub   string `yaml:"github,omitempty" json:"github,omitempty"`
	Telegram string `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	WeChat   string `yaml:"wechat,omitempty" json:"wechat,omitempty"`
	WhatsApp string `yaml:"whatsapp,omitempty" json:"whatsapp,omitempty"`
	X        string `yaml:"x,omitempty" json:"x,omitempty"`
}

// CommunityData represents user preferences for community participation.
// Email is strictly forbidden to prevent spam.
type CommunityData struct {
	ID           string           `yaml:"id,omitempty" json:"id,omitempty"`
	ArtisanToken string           `yaml:"artisan_token,omitempty" json:"artisan_token,omitempty"`
	Enabled      bool             `yaml:"enabled" json:"enabled"`
	Nickname     string           `yaml:"nickname,omitempty" json:"nickname,omitempty"`
	Location     string           `yaml:"location,omitempty" json:"location,omitempty"`
	SourceURL    string           `yaml:"source_url,omitempty" json:"source_url,omitempty"` // repository or project website
	Bio          string           `yaml:"bio,omitempty" json:"bio,omitempty"`               // max 200 characters
	Contact      string           `yaml:"contact,omitempty" json:"contact,omitempty"`       // primary public contact handle/URL (never email!)
	Socials      CommunitySocials `yaml:"socials,omitempty" json:"socials,omitempty"`
}

// Endpoint returns the active community API endpoint URL.
func Endpoint() string {
	if ep := strings.TrimSpace(os.Getenv("EGGS_COMMUNITY_ENDPOINT")); ep != "" {
		return ep
	}
	if ep := strings.TrimSpace(os.Getenv("CENSUS_COMMUNITY_URL")); ep != "" {
		return ep
	}
	return DefaultCommunityEndpoint
}
