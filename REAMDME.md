Presentation
L’objectif est de créer une application web utilisant une API externe, avec :
-affichage dynamique des données,
-filtres,
-page de détails,
-système de favoris.

L’application propose les fonctionnalités suivantes :

-Affichage de toutes les cartes Pokémon
-Recherche par nom ou identifiant
-Filtrage par type d’énergie
-Filtrage par série
-Page de détails pour chaque carte
-Système de favoris stocké en .JSON

-Langages utilisé:
-Go
-net/http
-html/template
-Css
-API : https://api.tcgdex.net

Arborescence du projet:

-main.go
-go.mod
-data/
   -favorites.json
-internal/
  -tcgdex/
  -favorites/
templates/
  -header.html
  -home.html
  -details.html
  -favorites.html
  -error.html
static/
  -style.css
  -app.js

Comment lancer le projet :

-Go 
-Git
-Connexion internet

Récuperer le projet :

git clone https://github.com/quentinfue/Projet-groupie-tracker-Fuentes-Quentin.git
go run .