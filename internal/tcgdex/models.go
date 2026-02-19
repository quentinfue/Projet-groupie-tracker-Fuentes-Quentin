package tcgdex

type Series struct {
	ID   string
	Name string
}

type CardLite struct {
	ID       string
	Name     string
	Image    string
	SetID    string
	LocalID  string
	SeriesID string
}

type Card struct {
	ID       string
	Name     string
	Image    string
	SetID    string
	LocalID  string
	HP       string
	Rarity   string
	Types    []string
	SeriesID string
}
