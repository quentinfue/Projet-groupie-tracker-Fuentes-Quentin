package tcgdex

type CardLite struct {
	ID       string
	Name     string
	Image    string
	SetID    string
	LocalID  string
	SeriesID string
	Rarity   string
	Types    []string
}

type CardDetails struct {
	ID      string
	Name    string
	Image   string
	HP      int
	Rarity  string
	Types   []string
	SetID   string
	LocalID string
}
