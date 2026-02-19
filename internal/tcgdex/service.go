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

type IDRef struct {
	ID string
}

func (r *IDRef) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		r.ID = s
		return nil
	}

	var obj struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	r.ID = obj.ID
	return nil
}

type Service struct {
	BaseURL string
	Client  *http.Client
}

func NewService(baseURL string) *Service {
	return &Service{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Service) getJSON(u string, out any) error {
	resp, err := s.Client.Get(u)
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

func fixIDs(id, setID, localID string) (string, string) {
	if setID == "" || localID == "" {
		parts := strings.SplitN(id, "-", 2)
		if len(parts) == 2 {
			if setID == "" {
				setID = parts[0]
			}
			if localID == "" {
				localID = parts[1]
			}
		}
	}
	return setID, localID
}

func (s *Service) ListAllCards() ([]CardLite, error) {
	endpoint := s.BaseURL + "/cards"

	var raw []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Image   string `json:"image"`
		Set     IDRef  `json:"set"`
		LocalID string `json:"localId"`
		Series  IDRef  `json:"serie"`
	}

	if err := s.getJSON(endpoint, &raw); err != nil {
		return nil, err
	}

	out := make([]CardLite, 0, len(raw))

	for _, c := range raw {
		setID, localID := fixIDs(c.ID, c.Set.ID, c.LocalID)

		out = append(out, CardLite{
			ID:       c.ID,
			Name:     c.Name,
			Image:    c.Image,
			SetID:    setID,
			LocalID:  localID,
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
		ID      string `json:"id"`
		Name    string `json:"name"`
		Image   string `json:"image"`
		Set     IDRef  `json:"set"`
		LocalID string `json:"localId"`
		Series  IDRef  `json:"serie"`
	}

	if err := s.getJSON(u.String(), &raw); err != nil {
		return nil, err
	}

	out := make([]CardLite, 0, len(raw))

	for _, c := range raw {
		setID, localID := fixIDs(c.ID, c.Set.ID, c.LocalID)

		out = append(out, CardLite{
			ID:       c.ID,
			Name:     c.Name,
			Image:    c.Image,
			SetID:    setID,
			LocalID:  localID,
			SeriesID: c.Series.ID,
		})
	}

	return out, nil
}

func (s *Service) GetCardByID(id string) (Card, error) {
	endpoint := fmt.Sprintf("%s/cards/%s", s.BaseURL, url.PathEscape(id))

	var raw struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Image   string   `json:"image"`
		Set     IDRef    `json:"set"`
		LocalID string   `json:"localId"`
		HP      any      `json:"hp"`
		Rarity  string   `json:"rarity"`
		Types   []string `json:"types"`
		Series  IDRef    `json:"serie"`
	}

	if err := s.getJSON(endpoint, &raw); err != nil {
		return Card{}, err
	}

	hp := ""

	switch v := raw.HP.(type) {
	case string:
		hp = v
	case float64:
		hp = fmt.Sprintf("%.0f", v)
	}

	setID, localID := fixIDs(raw.ID, raw.Set.ID, raw.LocalID)

	return Card{
		ID:       raw.ID,
		Name:     raw.Name,
		Image:    raw.Image,
		SetID:    setID,
		LocalID:  localID,
		HP:       hp,
		Rarity:   raw.Rarity,
		Types:    raw.Types,
		SeriesID: raw.Series.ID,
	}, nil
}

func (s *Service) ListSeries() ([]Series, error) {
	endpoint := s.BaseURL + "/series"

	var raw []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := s.getJSON(endpoint, &raw); err != nil {
		return nil, err
	}

	out := make([]Series, 0, len(raw))
	for _, it := range raw {
		out = append(out, Series{
			ID:   it.ID,
			Name: it.Name,
		})
	}

	return out, nil
}
