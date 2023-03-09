package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var db *sql.DB

func main() {
	var err error
	godotenv.Load("data.env")
	db, err = sql.Open("postgres", "postgresql:///frigid?sslmode=disable")
	if err != nil {
		log.Fatalln("error opening db: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln("error pinging db, ", err)
	}
	var recipe Recipe
	url := "https://tasty.p.rapidapi.com/recipes/list?from=160&size=40&tags=under_30_minutes"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("error making request to API: ", err)
	}
	req.Header.Add("X-RapidAPI-Key", os.Getenv("APIKEY"))
	req.Header.Add("X-RapidAPI-Host", "tasty.p.rapidapi.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("error receiving response from API: ", err)
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("error reading response from API :", err)
	}

	err = json.Unmarshal(bs, &recipe)
	if err != nil {
		log.Fatalln("error unmarshalling into recipes: ", err)
	}
	db.Query("CREATE TABLE recipes (recipe_id VARCHAR,  recipe_name VARCHAR, ingredients VARCHAR, description VARCHAR, instructions VARCHAR)")
	for _, results := range recipe.Results {
		var recipeName string
		var recipeID string
		var recipeDescription string
		var recipeInstruction []string
		var recipeIngredients []string
		recipeName = results.Name
		recipeID = strconv.Itoa(results.ID)
		recipeDescription = results.Description
		for _, value := range results.Instructions {
			recipeInstruction = append(recipeInstruction, value.DisplayText)
		}
		for _, components := range results.Sections {
			for _, ingredients := range components.Components {
				recipeIngredients = append(recipeIngredients, ingredients.Ingredient.Name)
			}
		}
		instructionstring := strings.Join(recipeInstruction, "\n")
		ingredientstring := strings.Join(recipeIngredients, "\n")
		fmt.Println(recipeName, recipeID, recipeDescription, ingredientstring, instructionstring)
		stmt, err := db.Prepare("INSERT INTO recipes (recipe_id, recipe_name, ingredients, description, instructions) VALUES ($1,$2,$3,$4,$5);")
		if err != nil {
			log.Fatalln("error preparing sql query:", err)
		}
		_, err = stmt.Exec(recipeID, recipeName, ingredientstring, recipeDescription, instructionstring)
		if err != nil {
			log.Fatalln("unable to execute query: ", err)
		}
		stmt.Close()
	}
}

type Recipe struct {
	Count   int `json:"count"`
	Results []struct {
		VideoID        int         `json:"video_id"`
		Keywords       string      `json:"keywords"`
		BuzzID         interface{} `json:"buzz_id"`
		CreatedAt      int         `json:"created_at"`
		UpdatedAt      int         `json:"updated_at"`
		IsShoppable    bool        `json:"is_shoppable"`
		VideoAdContent string      `json:"video_ad_content"`
		UserRatings    struct {
			CountNegative int     `json:"count_negative"`
			CountPositive int     `json:"count_positive"`
			Score         float64 `json:"score"`
		} `json:"user_ratings"`
		ID        int `json:"id"`
		Nutrition struct {
			Fat           int       `json:"fat"`
			Calories      int       `json:"calories"`
			Sugar         int       `json:"sugar"`
			Carbohydrates int       `json:"carbohydrates"`
			Fiber         int       `json:"fiber"`
			UpdatedAt     time.Time `json:"updated_at"`
			Protein       int       `json:"protein"`
		} `json:"nutrition"`
		Description         string        `json:"description"`
		DraftStatus         string        `json:"draft_status"`
		TotalTimeMinutes    interface{}   `json:"total_time_minutes"`
		Yields              string        `json:"yields"`
		NutritionVisibility string        `json:"nutrition_visibility"`
		Language            string        `json:"language"`
		BrandID             int           `json:"brand_id"`
		AspectRatio         string        `json:"aspect_ratio"`
		IsOneTop            bool          `json:"is_one_top"`
		BeautyURL           interface{}   `json:"beauty_url"`
		OriginalVideoURL    string        `json:"original_video_url"`
		Promotion           string        `json:"promotion"`
		FacebookPosts       []interface{} `json:"facebook_posts"`
		Sections            []struct {
			Components []struct {
				RawText      string `json:"raw_text"`
				ExtraComment string `json:"extra_comment"`
				Ingredient   struct {
					ID              int    `json:"id"`
					DisplaySingular string `json:"display_singular"`
					UpdatedAt       int    `json:"updated_at"`
					Name            string `json:"name"`
					CreatedAt       int    `json:"created_at"`
					DisplayPlural   string `json:"display_plural"`
				} `json:"ingredient"`
				ID           int `json:"id"`
				Position     int `json:"position"`
				Measurements []struct {
					Unit struct {
						DisplayPlural   string `json:"display_plural"`
						DisplaySingular string `json:"display_singular"`
						Abbreviation    string `json:"abbreviation"`
						System          string `json:"system"`
						Name            string `json:"name"`
					} `json:"unit"`
					Quantity string `json:"quantity"`
					ID       int    `json:"id"`
				} `json:"measurements"`
			} `json:"components"`
			Name     string `json:"name"`
			Position int    `json:"position"`
		} `json:"sections"`
		Name string `json:"name"`
		Show struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"show"`
		Renditions []struct {
			Duration       int         `json:"duration"`
			ContentType    string      `json:"content_type"`
			Aspect         string      `json:"aspect"`
			MinimumBitRate interface{} `json:"minimum_bit_rate"`
			MaximumBitRate interface{} `json:"maximum_bit_rate"`
			FileSize       int         `json:"file_size"`
			URL            string      `json:"url"`
			BitRate        int         `json:"bit_rate"`
			Width          int         `json:"width"`
			Name           string      `json:"name"`
			Height         int         `json:"height"`
			Container      string      `json:"container"`
			PosterURL      string      `json:"poster_url"`
		} `json:"renditions"`
		TotalTimeTier interface{} `json:"total_time_tier"`
		Instructions  []struct {
			ID          int         `json:"id"`
			Position    int         `json:"position"`
			DisplayText string      `json:"display_text"`
			StartTime   int         `json:"start_time"`
			Appliance   interface{} `json:"appliance"`
			EndTime     int         `json:"end_time"`
			Temperature interface{} `json:"temperature"`
		} `json:"instructions"`
		TipsAndRatingsEnabled bool        `json:"tips_and_ratings_enabled"`
		InspiredByURL         interface{} `json:"inspired_by_url"`
		ServingsNounPlural    string      `json:"servings_noun_plural"`
		Topics                []struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"topics"`
		Brand struct {
			ImageURL string `json:"image_url"`
			Name     string `json:"name"`
			ID       int    `json:"id"`
			Slug     string `json:"slug"`
		} `json:"brand"`
		Slug string `json:"slug"`
		Tags []struct {
			ID          int    `json:"id"`
			DisplayName string `json:"display_name"`
			Type        string `json:"type"`
			Name        string `json:"name"`
		} `json:"tags"`
		NumServings      int    `json:"num_servings"`
		ThumbnailAltText string `json:"thumbnail_alt_text"`
		Credits          []struct {
			ImageURL string `json:"image_url"`
			Name     string `json:"name"`
			ID       int    `json:"id"`
			Type     string `json:"type"`
			Slug     string `json:"slug"`
		} `json:"credits"`
		Price struct {
			Total              int       `json:"total"`
			UpdatedAt          time.Time `json:"updated_at"`
			Portion            int       `json:"portion"`
			ConsumptionTotal   int       `json:"consumption_total"`
			ConsumptionPortion int       `json:"consumption_portion"`
		} `json:"price"`
		ShowID               int           `json:"show_id"`
		PrepTimeMinutes      int           `json:"prep_time_minutes"`
		ThumbnailURL         string        `json:"thumbnail_url"`
		CanonicalID          string        `json:"canonical_id"`
		CookTimeMinutes      int           `json:"cook_time_minutes"`
		Country              string        `json:"country"`
		ServingsNounSingular string        `json:"servings_noun_singular"`
		Compilations         []interface{} `json:"compilations"`
		VideoURL             string        `json:"video_url"`
		ApprovedAt           int           `json:"approved_at"`
		SeoTitle             string        `json:"seo_title"`
	} `json:"results"`
}
