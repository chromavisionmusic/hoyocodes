package hoyo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GameInfo struct {
	ID        string
	Name      string
	Color     int
	RedeemURL string
	Thumbnail string
}

var Games = map[string]GameInfo{
	"genshin": {
		ID:        "genshin",
		Name:      "Genshin Impact",
		Color:     0x4BA3E3, // Anemo/Hydro cyan-blue
		RedeemURL: "https://genshin.hoyoverse.com/en/gift?code=%s",
		Thumbnail: "https://static.wikia.nocookie.net/logopedia/images/9/91/Genshin_Impact_Icon.png/revision/latest?cb=20241230075438",
	},
	"hkrpg": {
		ID:        "hkrpg",
		Name:      "Honkai: Star Rail",
		Color:     0xD0A040, // Astral Express Gold
		RedeemURL: "https://hsr.hoyoverse.com/gift?code=%s",
		Thumbnail: "https://upload-os-bbs.hoyolab.com/upload/2025/08/15/63353547/e067445d62d978d7c9792d9043433365_2554893110477649291.webp",
	},
	"nap": {
		ID:        "nap",
		Name:      "Zenless Zone Zero",
		Color:     0xFF5A00, // Vibrant Bangboo Orange
		RedeemURL: "https://zenless.hoyoverse.com/redemption?code=%s",
		Thumbnail: "https://static.wikia.nocookie.net/logopedia/images/d/df/Zenless_Zone_Zero_%28Icon_without_any_logo%29.png/revision/latest?cb=20250411003400",
	},
}

func GetGameInfo(game string) (GameInfo, bool) {
	info, ok := Games[game]
	return info, ok
}

type CodeItem struct {
	ID      int    `json:"id"`
	Code    string `json:"code"`
	Status  string `json:"status"`
	Game    string `json:"game"`
	Rewards string `json:"rewards"`
}

type APIResponse struct {
	Codes []CodeItem `json:"codes"`
	Game  string     `json:"game"`
}

type cacheEntry struct {
	codes     []CodeItem
	expiresAt time.Time
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	cacheTTL   time.Duration
	cache      map[string]cacheEntry
	mu         sync.RWMutex
}

func NewClient(baseURL string, ttl time.Duration) *Client {
	if baseURL == "" {
		baseURL = "https://hoyo-codes.seria.moe"
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: ttl,
		cache:    make(map[string]cacheEntry),
	}
}

var DefaultClient = NewClient("https://hoyo-codes.seria.moe", 5*time.Minute)

func (c *Client) GetCodes(ctx context.Context, game string) ([]CodeItem, error) {
	c.mu.RLock()
	entry, found := c.cache[game]
	c.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		return entry.codes, nil
	}

	reqURL := fmt.Sprintf("%s/codes?game=%s", c.baseURL, game)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "HoyoCodes-Discord-Bot/0.1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// If cached data is available even if expired, return it as fallback during network failure
		if found && len(entry.codes) > 0 {
			return entry.codes, nil
		}
		return nil, fmt.Errorf("fetching codes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if found && len(entry.codes) > 0 {
			return entry.codes, nil
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		if found && len(entry.codes) > 0 {
			return entry.codes, nil
		}
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.mu.Lock()
	c.cache[game] = cacheEntry{
		codes:     apiResp.Codes,
		expiresAt: time.Now().Add(c.cacheTTL),
	}
	c.mu.Unlock()

	return apiResp.Codes, nil
}

func (g GameInfo) FormatRedeemURL(code string) string {
	if g.RedeemURL == "" {
		return ""
	}
	return fmt.Sprintf(g.RedeemURL, code)
}
