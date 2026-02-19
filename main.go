package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"Projet-groupie-tracker-Fuentes-Quentin/internal/favorites"
	"Projet-groupie-tracker-Fuentes-Quentin/internal/tcgdex"
)

type App struct {
	TPL *template.Template
	API *tcgdex.Service
	Fav *favorites.Store
}

func main() {
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}).ParseFiles(
		"templates/header.html",
		"templates/home.html",
		"templates/details.html",
		"templates/favorites.html",
		"templates/error.html",
	))

	app := &App{
		TPL: tpl,
		API: tcgdex.NewService("https://api.tcgdex.net/v2/fr"),
		Fav: favorites.NewStore("data/favorites.json"),
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", app.homeHandler)
	mux.HandleFunc("/pokemon/details", app.detailsHandler)
	mux.HandleFunc("/favorites", app.favoritesHandler)
	mux.HandleFunc("/api/favorites/toggle", app.toggleFavoriteHandler)

	log.Println("Server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (a *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := strings.TrimSpace(r.URL.Query().Get("type"))
	series := strings.TrimSpace(r.URL.Query().Get("series"))
	page := parseIntDefault(r.URL.Query().Get("page"), 1)

	perPage := 20

	seriesList, _ := a.API.ListSeries()

	var cards []tcgdex.CardLite
	var err error

	if typ != "" {
		cards, err = a.API.ListCardsByType(typ)
	} else {
		cards, err = a.API.ListAllCards()
	}

	if err != nil {
		a.render(w, "error.html", map[string]any{
			"Title":   "Erreur",
			"Code":    503,
			"Message": "Erreur API",
		})
		return
	}

	filtered := make([]tcgdex.CardLite, 0)

	for _, c := range cards {

		if q != "" {
			if !strings.Contains(strings.ToLower(c.Name), strings.ToLower(q)) &&
				!strings.Contains(strings.ToLower(c.ID), strings.ToLower(q)) {
				continue
			}
		}

		if series != "" && c.SeriesID != series {
			continue
		}

		filtered = append(filtered, c)
	}

	total := len(filtered)
	totalPages := (total + perPage - 1) / perPage

	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		start = total
	}

	if end > total {
		end = total
	}

	data := map[string]any{
		"Title":          "Accueil",
		"Cards":          filtered[start:end],
		"Favs":           a.Fav.AllSet(),
		"Q":              q,
		"Type":           typ,
		"Series":         seriesList,
		"SeriesSelected": series,
		"Page":           page,
		"TotalPages":     totalPages,
		"Total":          total,
	}

	a.render(w, "home.html", data)
}

func (a *App) detailsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	setID := strings.TrimSpace(r.URL.Query().Get("set"))
	localID := strings.TrimSpace(r.URL.Query().Get("local"))

	var card tcgdex.Card
	var err error

	if id != "" {
		card, err = a.API.GetCardByID(id)
	} else {
		if setID == "" || localID == "" {
			a.render(w, "error.html", map[string]any{
				"Title":   "Erreur",
				"Code":    400,
				"Message": "Paramètres manquants",
			})
			return
		}

		card, err = a.API.GetCardByID(setID + "-" + localID)
	}

	if err != nil {
		a.render(w, "error.html", map[string]any{
			"Title":   "Erreur",
			"Code":    503,
			"Message": "Carte introuvable",
		})
		return
	}

	a.render(w, "details.html", map[string]any{
		"Title": "Détails",
		"Card":  card,
		"Favs":  a.Fav.AllSet(),
	})
}

func (a *App) favoritesHandler(w http.ResponseWriter, r *http.Request) {
	favIDs := a.Fav.All()
	setFav := map[string]bool{}
	for _, id := range favIDs {
		setFav[id] = true
	}

	cards, err := a.API.ListAllCards()
	if err != nil {
		a.render(w, "error.html", map[string]any{
			"Title":   "Erreur",
			"Code":    503,
			"Message": "Erreur API",
		})
		return
	}

	out := make([]tcgdex.CardLite, 0)
	for _, c := range cards {
		if setFav[c.ID] {
			out = append(out, c)
		}
	}

	a.render(w, "favorites.html", map[string]any{
		"Title": "Favoris",
		"Cards": out,
		"Favs":  a.Fav.AllSet(),
	})
}

func (a *App) toggleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", 400)
		return
	}

	id := strings.TrimSpace(r.FormValue("id"))
	next := strings.TrimSpace(r.FormValue("next"))

	if id == "" {
		http.Error(w, "Missing id", 400)
		return
	}

	_, err := a.Fav.Toggle(id)
	if err != nil {
		http.Error(w, "Save error", 500)
		return
	}

	if next == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}

	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (a *App) render(w http.ResponseWriter, name string, data map[string]any) {
	if err := a.TPL.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Template error", 500)
	}
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
