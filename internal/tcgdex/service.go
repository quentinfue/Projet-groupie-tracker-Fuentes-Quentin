package tcgdex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Service struct {
	BaseURL string
	Client  *http.Client
}

func NewService(baseURL string) *Service {
	return &Service{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *Service) ListAllCards() ([]CardLite, error) {
	endpoint := s.BaseURL + "/cards"
	var raw []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Image string `json:"image"`
		Set   struct {
			ID string `json:"id"`
		} `json:"set"`
		LocalID string `json:"localId"`
		Series  struct {
			ID string `json:"id"`
		} `json:"serie"`
	}

	if err := s.getJSON(endpoint, &raw); err != nil {
		return nil, err
	}

	out := make([]CardLite, 0, len(raw))
	for _, c := range raw {
		out = append(out, CardLite{
			ID:       c.ID,
			Name:     c.Name,
			Image:    c.Image,
			SetID:    c.Set.ID,
			LocalID:  c.LocalID,
			SeriesID: c.Series.ID,
		})
	}
	return out, nil
}

func (s *Service) ListCardsByType(typ string) ([]CardLite, error) {
	u, _ := url.Parse(s.BaseURL + "/cards")
	q := u.Query()
	q.Set("types", typ)
	u.RawQuery = q.Encode()

	var raw []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Image string `json:"image"`
		Set   struct {
			ID string `json:"id"`
		} `json:"set"`
		LocalID string `json:"localId"`
		Series  struct {
			ID string `json:"id"`
		} `json:"serie"`
	}

	if err := s.getJSON(u.String(), &raw); err != nil {
		return nil, err
	}

	out := make([]CardLite, 0, len(raw))
	for _, c := range raw {
		out = append(out, CardLite{
			ID:       c.ID,
			Name:     c.Name,
			Image:    c.Image,
			SetID:    c.Set.ID,
			LocalID:  c.LocalID,
			SeriesID: c.Series.ID,
		})
	}
	return out, nil
}

// Détails via /sets/{set}/{local}
func (s *Service) GetCardFromSet(setID, localID string) (Card, error) {
	endpoint := fmt.Sprintf("%s/sets/%s/%s", s.BaseURL, url.PathEscape(setID), url.PathEscape(localID))

	var raw struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Image string `json:"image"`
		Set   struct {
			ID string `json:"id"`
		} `json:"set"`
		LocalID string   `json:"localId"`
		HP      string   `json:"hp"`
		Rarity  string   `json:"rarity"`
		Types   []string `json:"types"`
		Series  struct {
			ID string `json:"id"`
		} `json:"serie"`
	}

	if err := s.getJSON(endpoint, &raw); err != nil {
		return Card{}, err
	}

	return Card{
		ID:       raw.ID,
		Name:     raw.Name,
		Image:    raw.Image,
		SetID:    raw.Set.ID,
		LocalID:  raw.LocalID,
		HP:       raw.HP,
		Rarity:   raw.Rarity,
		Types:    raw.Types,
		SeriesID: raw.Series.ID,
	}, nil
}

func (s *Service) getJSON(url string, out any) error {
	resp, err := s.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tcgdex http %d: %s", resp.StatusCode, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
