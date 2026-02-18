package tcgdex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	BaseURL string
	Client  *http.Client
}

func NewService(base string) *Service {
	return &Service{
		BaseURL: base,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) ListAllCards() ([]CardLite, error) {
	url := s.BaseURL + "/cards"
	var raw []map[string]any
	if err := s.getJSON(url, &raw); err != nil {
		return nil, err
	}
	return mapCardsLite(raw), nil
}

func (s *Service) ListCardsByType(typ string) ([]CardLite, error) {
	url := fmt.Sprintf("%s/cards?types=%s", s.BaseURL, typ)
	var raw []map[string]any
	if err := s.getJSON(url, &raw); err != nil {
		return nil, err
	}
	return mapCardsLite(raw), nil
}

func (s *Service) GetCardFromSet(setID, localID string) (CardDetails, error) {
	url := fmt.Sprintf("%s/sets/%s/%s", s.BaseURL, setID, localID)
	var raw map[string]any
	if err := s.getJSON(url, &raw); err != nil {
		return CardDetails{}, err
	}
	return mapCardDetails(raw), nil
}

func (s *Service) getJSON(url string, out any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func mapCardsLite(raw []map[string]any) []CardLite {
	cards := make([]CardLite, 0, len(raw))
	for _, r := range raw {
		var c CardLite

		if v, ok := r["id"].(string); ok {
			c.ID = v
		}
		if v, ok := r["name"].(string); ok {
			c.Name = v
		}
		if v, ok := r["localId"].(string); ok {
			c.LocalID = v
		}
		if v, ok := r["image"].(string); ok {
			c.Image = v
		}

		if c.ID != "" {
			parts := strings.SplitN(c.ID, "-", 2)
			if len(parts) == 2 {
				c.SetID = parts[0]
				if c.LocalID == "" {
					c.LocalID = parts[1]
				}
			}
		}

		c.SeriesID = extractSeriesID(c.SetID)

		if c.ID != "" && c.Name != "" {
			cards = append(cards, c)
		}
	}
	return cards
}

func extractSeriesID(setID string) string {
	if setID == "" {
		return ""
	}
	if i := strings.Index(setID, "-"); i > 0 {
		return strings.ToLower(setID[:i])
	}
	var b strings.Builder
	for _, r := range setID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		} else {
			break
		}
	}
	return strings.ToLower(b.String())
}

func mapCardDetails(r map[string]any) CardDetails {
	var c CardDetails

	if v, ok := r["id"].(string); ok {
		c.ID = v
	}
	if v, ok := r["name"].(string); ok {
		c.Name = v
	}
	if v, ok := r["image"].(string); ok {
		c.Image = v
	}

	switch hp := r["hp"].(type) {
	case float64:
		c.HP = int(hp)
	case string:
		hp = strings.TrimSpace(hp)
		if hp != "" {
			var n int
			_, _ = fmt.Sscanf(hp, "%d", &n)
			c.HP = n
		}
	}

	if v, ok := r["rarity"].(string); ok {
		c.Rarity = v
	}

	if arr, ok := r["types"].([]any); ok {
		for _, it := range arr {
			if s, ok := it.(string); ok {
				c.Types = append(c.Types, s)
			}
		}
	}

	if set, ok := r["set"].(map[string]any); ok {
		if id, ok := set["id"].(string); ok {
			c.SetID = id
		}
	}

	if v, ok := r["localId"].(string); ok {
		c.LocalID = v
	}

	if c.ID != "" && (c.SetID == "" || c.LocalID == "") {
		parts := strings.SplitN(c.ID, "-", 2)
		if len(parts) == 2 {
			if c.SetID == "" {
				c.SetID = parts[0]
			}
			if c.LocalID == "" {
				c.LocalID = parts[1]
			}
		}
	}

	return c
}
