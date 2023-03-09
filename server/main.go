package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var db *sql.DB

type Recipe struct {
	Recipe_Id    int    `json:"recipe_id"`
	Recipe_Name  string `json:"recipe_name"`
	Ingredients  string `json:"ingredients"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

type User struct {
	User_id  int    `json:"user_id"`
	Username string `json:"username"`
}

type Rating struct {
	Rating_id int  `json:"rating_id"`
	User_id   int  `json:"user_id"`
	Recipe_id int  `json:"recipe_id"`
	Rating    bool `json:"rating"`
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgresql:///frigid?sslmode=disable")
	if err != nil {
		log.Fatalln("error opening db: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln("error pinging db, ", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/recipelist", GetRecipes).Methods("GET")
	r.HandleFunc("/ingredients", SearchByIngredients).Methods("GET")
	r.HandleFunc("/recipe", SearchByID).Methods("GET")
	r.HandleFunc("/create", CreateNewRecipe).Methods("POST")
	r.HandleFunc("/delete", DeleteRecipe).Methods("DELETE")
	r.HandleFunc("/update", UpdateRecipe).Methods("PUT")
	r.HandleFunc("/createuser", CreateUser).Methods("POST")
	r.HandleFunc("/createrating", CreateRating).Methods("POST")
	r.HandleFunc("/getrating/{id}", GetRatingbyUID).Methods("GET")
	r.HandleFunc("/getratingbyrecipe/{id)", GetRatingsbyRecipe).Methods("GET")
	r.HandleFunc("/user", GetUser).Methods("GET")
	r.HandleFunc("/deleterating/{id}", DeleteRating).Methods("DELETE")
	fmt.Println("Server is running")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}

func GetRecipes(w http.ResponseWriter, r *http.Request) {
	var recipes []Recipe
	limit := r.FormValue("limit")
	stmt, err := db.Prepare("SELECT * FROM recipes LIMIT $1;")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	result, err := stmt.Query(limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = stmt.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for result.Next() {
		var recipe Recipe
		err = result.Scan(&recipe.Recipe_Id, &recipe.Recipe_Name, &recipe.Ingredients, &recipe.Description, &recipe.Instructions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		recipes = append(recipes, recipe)
	}
	bs, err := json.Marshal(recipes)
	if err != nil {
		log.Fatalln("error marshalling recipes: ", err)
	}
	_, err = w.Write(bs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func SearchByIngredients(w http.ResponseWriter, r *http.Request) {
	var recipes []Recipe
	search := r.FormValue("search")
	search = search2query(search)
	stmt, err := db.Prepare("select * from recipes where ingredients LIKE '%' ||$1|| '%'")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	result, err := stmt.Query(search)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	err = stmt.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for result.Next() {
		var recipe Recipe
		err = result.Scan(&recipe.Recipe_Id, &recipe.Recipe_Name, &recipe.Ingredients, &recipe.Description, &recipe.Instructions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		recipes = append(recipes, recipe)
	}
	bs, err := json.Marshal(recipes)
	if err != nil {
		log.Fatalln("error marshalling recipes: ", err)
	}
	_, err = w.Write(bs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
func search2query(search string) (query string) {
	split := strings.Split(search, ",")
	newsearch := strings.Join(split, "%")
	query = "%" + newsearch
	return query
}

func SearchByID(w http.ResponseWriter, r *http.Request) {
	var recipe Recipe
	id := r.FormValue("id")
	stmt, err := db.Prepare("SELECT * FROM recipes WHERE recipe_id = $1;")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	intid, err := strconv.Atoi(id)
	results := stmt.QueryRow(intid)
	err = stmt.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	err = results.Scan(&recipe.Recipe_Id, &recipe.Recipe_Name, &recipe.Description, &recipe.Ingredients, &recipe.Instructions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	bs, err := json.Marshal(recipe)
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
	_, err = w.Write(bs)
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
}
func CreateNewRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe Recipe
	request, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalln(err)
		return
	}
	err = json.Unmarshal(request, &recipe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalln(err)
		return
	}

	var newID int
	err = db.QueryRow("SELECT MAX(recipe_id) FROM recipes").Scan(&newID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		log.Fatalln("Error: Failed to create new recipe")
		return
	}

	newID = newID + 1
	stmt, err := db.Prepare("INSERT INTO recipes(recipe_name,ingredients,description,instructions, recipe_id) VALUES ($1, $2, $3, $4, $5) ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
	}
	defer stmt.Close()
	result, err := stmt.Query(recipe.Recipe_Name, recipe.Ingredients, recipe.Description, recipe.Instructions, newID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	err = result.Close()
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(recipe)
}

func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	stmt, err := db.Prepare("DELETE FROM recipes WHERE recipe_id = $1 ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
	}
	result, err := stmt.Exec(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error: Failed to delete recipe")
		return
	}
	if rowsAffected <= 0 {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("error: Recipe matching that ID cannot be found")
	}
	w.WriteHeader(200)
}

func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	request, err := io.ReadAll(r.Body)
	var recipe Recipe
	json.Unmarshal(request, &recipe)
	stmt, err := db.Prepare("UPDATE recipes SET recipe_name = $2, ingredients = $3, description = $4, instructions = $5  WHERE recipe_id = $1")
	rows, err := stmt.Exec(recipe.Recipe_Id, recipe.Recipe_Name, recipe.Ingredients, recipe.Description, recipe.Instructions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	rowsaffected, err := rows.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	if rowsaffected <= 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Error: No recipe matching given recipe ID can be found")
	}
	w.WriteHeader(200)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	request, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(request, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Println("error decoding JSON from user")
		log.Fatalln(err)
	}
	var newID int
	err = db.QueryRow("SELECT MAX(user_id) FROM users").Scan(&newID)
	if err != nil {
		log.Println("error querying db to find max")
		log.Fatalln(err)
	}
	newID = newID + 1
	stmt, err := db.Prepare("INSERT INTO users(user_id, username) VALUES ($1, $2)")
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newID, user.Username)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		log.Fatalln("unable to create new user")
	}
	w.WriteHeader(http.StatusCreated)
}

func CreateRating(w http.ResponseWriter, r *http.Request) {
	var rating Rating
	request, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(request, &rating)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Unable to unmarshal request body into rating")
		log.Fatalln(err)
	}
	var newID int
	err = db.QueryRow("SELECT MAX(rating_id) from ratings").Scan(&newID)
	if err != nil {
		log.Println("error querying db to create rating ID")
		log.Println(err)
		w.WriteHeader(500)
	}
	newID = newID + 1
	stmt, err := db.Prepare("INSERT INTO ratings(rating_id, user_id, recipe_id, rating) VALUES ($1, $2, $3, $4)")
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newID, rating.User_id, rating.Recipe_id, rating.Rating)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
	}
	w.WriteHeader(200)
}

func GetRatingbyUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var ratings []Rating
	stringid := vars["id"]
	intid, err := strconv.Atoi(stringid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	stmt, err := db.Prepare("SELECT * FROM ratings WHERE user_id = $1;")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	results, err := stmt.Query(intid)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
	}
	defer results.Close()
	for results.Next() {
		var rating Rating
		err = results.Scan(&rating.Rating_id, &rating.User_id, &rating.Recipe_id, &rating.Rating)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		ratings = append(ratings, rating)
	}
	bs, err := json.Marshal(ratings)
	if err != nil {
		log.Fatalln("error marshalling recipes: ", err)
	}
	_, err = w.Write(bs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetRatingsbyRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var ratings []Rating
	stringid := vars["id"]
	intid, err := strconv.Atoi(stringid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	stmt, err := db.Prepare("SELECT * FROM ratings WHERE recipe_id = $1;")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	results, err := stmt.Query(intid)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
	}
	defer results.Close()
	for results.Next() {
		var rating Rating
		err = results.Scan(&rating.Rating_id, &rating.User_id, &rating.Recipe_id, &rating.Rating)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		ratings = append(ratings, rating)
	}
	bs, err := json.Marshal(ratings)
	if err != nil {
		log.Fatalln("error marshalling recipes: ", err)
	}
	_, err = w.Write(bs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var user User
	username := r.FormValue("name")
	row := db.QueryRow("SELECT * FROM users WHERE username = $1", username)
	err := row.Scan(&user.User_id, &user.Username)
	if err != nil {
		log.Println("error scanning rows into struct user")
		log.Println(err)
		w.WriteHeader(500)
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(500)
		log.Fatalln("error encoding user")
	}
}

func DeleteRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	stmt, err := db.Prepare("DELETE FROM ratings WHERE rating_id = $1 ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
	}
	result, err := stmt.Exec(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error: Failed to delete rating")
		return
	}
	if rowsAffected <= 0 {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("error: rating matching that ID cannot be found")
	}
	w.WriteHeader(200)
}
