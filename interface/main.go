package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

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
	var nameEntry string
	fmt.Println("Welcome to Frigid!", "Please enter your username or if you do not have one, type 'create'")
	for {
		_, err := fmt.Scanln(&nameEntry)
		if err != nil {
			log.Fatalln("could not scan entry")
		}
		if nameEntry == "create" {
			name := register()
			nameEntry = name
			break
		}
		user := GetUser(nameEntry)
		if user.User_id <= 0 {
			fmt.Println("Sorry. That username is not in our system. If you'd like to create a profile, enter `create`")
		}
		if user.User_id >= 1 {
			break
		}
	}
	fmt.Printf("Welcome %v, please choose from the following options \n", nameEntry)
	var input string
	for {
		fmt.Println("To search for recipes by ingredients you want to use, enter 'search'")
		fmt.Println("To like a recipe, enter 'like'")
		fmt.Println("To view recipes you've liked, enter 'mylikes'")
		fmt.Println("If you'd like to remove a recipe from your likes, enter 'remove'")
		fmt.Println("To upload your own recipe, enter 'upload'")
		fmt.Println("See a typo or some incomplete information in a recipe? enter 'update' to fix it.")
		fmt.Println("To exit the program, enter 'exit'")
		fmt.Scanln(&input)
		if input == "search" {
			fmt.Println("Enter the ingredients you would like to use for your recipe. Use commas to separate your ingredients.")
			var ingredients string
			fmt.Scanln(&ingredients)
			GetRecipeByIngredients(ingredients)
		}
		if input == "like" {
			fmt.Println("Please enter the recipe ID of the recipe you would like to add to your likes")
			var id string
			fmt.Scanln(&id)
			u := GetUser(nameEntry)
			intid, err := strconv.Atoi(id)
			if err != nil {
				log.Println("Unable to scan Recipe ID")
				log.Fatalln(err)
			}
			LikeRecipe(intid, u)
			fmt.Println("You liked Recipe # :", intid)
		}
		if input == "mylikes" {
			u := GetUser(nameEntry)
			ratings := GetLikedRecipes(u)
			for _, rating := range ratings {
				recipe := GetRecipeByID(rating.Recipe_id)
				fmt.Println("Rating ID:", rating.Rating_id, "Recipe ID: ", recipe.Recipe_Id, "Recipe Name:", recipe.Recipe_Name)
			}
			for {
				var rID string
				fmt.Println("If you would like to see the instructions and ingredients for a recipe, enter the recipe_id number next to the recipe. If you wish to go back to the main menu, enter 'home'")
				fmt.Scanln(&rID)
				if rID == "home" {
					break
				}
				rIDstring, err := strconv.Atoi(rID)
				if err != nil {
					log.Println("error: invalid user input")
					log.Fatalln(err)
				}
				var found bool
				for _, rating := range ratings {
					if rIDstring == rating.Recipe_id {
						recipe := GetRecipeByID(rating.Recipe_id)
						fmt.Println(recipe.Recipe_Id, recipe.Recipe_Name)
						fmt.Println(recipe.Description)
						fmt.Println(recipe.Ingredients)
						fmt.Println(recipe.Instructions)
						found = true
					}
				}
				if found == false {
					fmt.Println("Unable to locate recipe with that ID in your liked recipes. Please make sure you are entering your recipe ID correctly.")
				}
			}
		}
		if input == "remove" {
			var removethisID string
			fmt.Println("Please enter the recipe ID of the recipe you wish to remove from your liked recipes")
			fmt.Scanln(&removethisID)
			intid, err := strconv.Atoi(removethisID)
			if err != nil {
				log.Println("unable to convert ID into int")
				log.Fatalln(err)
			}
			DeleteRating(intid)
			fmt.Println("Rating associated with rating ID: #", intid, " deleted.")
		}
		if input == "upload" {
			reader := bufio.NewReader(os.Stdin)
			var recipe Recipe
			fmt.Println("Welcome to the recipe builder!")
			fmt.Println("First name your recipe.")
			name, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Unable to read string")
				log.Fatalln(err)
			}
			recipe.Recipe_Name = name
			fmt.Println("You named your recipe:", recipe.Recipe_Name)
			fmt.Println("Next, enter a description of your dish")
			description, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Unable to read string")
				log.Fatalln(err)
			}
			recipe.Description = description
			fmt.Println("Next, enter all the ingredients needed for your dish, separated by commas")
			ingredients, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Unable to read string")
				log.Fatalln(err)
			}
			recipe.Ingredients = ingredients
			fmt.Println("Next, enter all the steps required to make your dish")
			instructions, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Unable to read string")
				log.Fatalln(err)
			}
			recipe.Instructions = instructions
			recipe.Recipe_Id = 1
			fmt.Println("Uploading Recipe to database.")
			CreateRecipe(recipe)
			fmt.Println("Your recipe :", recipe.Recipe_Name, "has been added to the database.")

		}
		if input == "update" {
			var recipe Recipe
			fmt.Println("Thanks for helping out! First, please enter the recipe ID for the recipe you would like to update.")
			var recipeID string
			for {
				_, err := fmt.Scanln(&recipeID)
				if err != nil {
					log.Println("Unable to scan recipe ID")
					log.Fatalln(err)
				}
				recipeIDint, err := strconv.Atoi(recipeID)
				if err != nil {
					log.Fatalln(err)
				}
				recipe = GetRecipeByID(recipeIDint)
				if recipe.Recipe_Id <= 0 {
					fmt.Println("Unable to locate a recipe with that recipe ID. Try again.")
					continue
				}
				if recipeID == "exit" {
					break
				}
				break
			}
			if recipeID == "exit" {
				break
			}
			var issue string
			fmt.Println("Great! Here's the recipe you requested. Which component of the recipe has an issue?")
			for {
				fmt.Println(recipe.Recipe_Name)
				fmt.Println("------------------")
				fmt.Println("If there's a typo above, enter 'name'")
				fmt.Println("------------------")
				fmt.Println(recipe.Description)
				fmt.Println("------------------")
				fmt.Println("If there's a typo above enter 'description'")
				fmt.Println("------------------")
				fmt.Println(recipe.Ingredients)
				fmt.Println("------------------")
				fmt.Println("If there's a typo above, enter 'ingredients'")
				fmt.Println("------------------")
				fmt.Println(recipe.Instructions)
				fmt.Println("if there's a typo above enter 'instructions'")
				fmt.Println("------------------")
				fmt.Println(" when you're finished editing this recipe, enter 'exit'")
				_, err := fmt.Scanln(&issue)
				if err != nil {
					log.Fatalln(err)
				}
				if issue == "name" {
					fmt.Println("Please enter the corrected name for this recipe")
					reader := bufio.NewReader(os.Stdin)
					newname, err := reader.ReadString('\n')
					if err != nil {
						log.Fatalln(err)
					}
					recipe.Recipe_Name = newname
				}
				if issue == "description" {
					fmt.Println("Please enter a new, corrected description for this recipe")
					reader := bufio.NewReader(os.Stdin)
					newdesc, err := reader.ReadString('\n')
					if err != nil {
						log.Fatalln(err)
					}
					recipe.Description = newdesc
				}
				if issue == "ingredients" {
					fmt.Println("Please enter the new list of ingredients separated by commas")
					reader := bufio.NewReader(os.Stdin)
					ingredientslist, err := reader.ReadString('\n')
					if err != nil {
						log.Fatalln(err)
					}
					recipe.Ingredients = ingredientslist
				}
				if issue == "instructions" {
					fmt.Println("Please enter the corrected instructions for this recipe")
					reader := bufio.NewReader(os.Stdin)
					instructions, err := reader.ReadString('\n')
					if err != nil {
						log.Fatalln(err)
					}
					recipe.Instructions = instructions
				}
				if issue == "exit" {
					url := "http://localhost:8080/update"
					jsonRecipe, err := json.Marshal(recipe)
					if err != nil {
						panic(err)
					}
					req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonRecipe))
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					log.Println(resp.StatusCode)
					break
				}
			}
		}
		if input == "exit" {
			break
		}
	}
}

func register() string {
	var username string
	var confirmedusername string
	for {
		fmt.Println("please enter what you would like your username to be:")
		_, err := fmt.Scanln(&username)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Great! You chose your username to be %v. Please confirm that you want this to be your username by entering it again.\n", username)
		fmt.Scanln(&confirmedusername)
		if username == confirmedusername {
			break
		}
		if username != confirmedusername {
			fmt.Println("Sorry those do not line up. Please confirm the username you would like to use")
		}
	}
	url := "http://localhost:8080/createuser"

	user := User{Username: username,
		User_id: 1}
	log.Println(user)
	jsonStr, err := json.Marshal(user)
	if err != nil {
		log.Println("error marshalling user")
		log.Fatalln(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("error sending post request")
		log.Fatalln(err)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	log.Println(resp.StatusCode)
	return user.Username
}
func GetUser(username string) User {
	url := "http://localhost:8080/user?name=" + username
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Fatalln("error: unable to decode response body: ", err)
	}
	return user
}
func GetRecipeByIngredients(ingredients string) {
	url := "http://localhost:8080/ingredients?search=" + ingredients
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var recipes []Recipe
	err = json.NewDecoder(resp.Body).Decode(&recipes)
	if err != nil {
		log.Fatalln("error decoding http response body: ", err)
	}
	for _, recipe := range recipes {
		fmt.Println("ID #: ", recipe.Recipe_Id)
		fmt.Println("Recipe Name: ", recipe.Recipe_Name)
		fmt.Println("Description:\n", recipe.Description)
		fmt.Println("Ingredients:\n", recipe.Ingredients)
		fmt.Println("Instructions:\n", recipe.Instructions)
		fmt.Println("-----------------")
	}
}
func LikeRecipe(id int, u User) {
	var userid int
	userid = u.User_id
	url := "http://localhost:8080/createrating"
	rating := Rating{
		Rating_id: 1,
		User_id:   userid,
		Recipe_id: id,
		Rating:    true,
	}
	jsonStr, err := json.Marshal(rating)
	if err != nil {
		log.Println("error marshalling rating")
		log.Fatalln(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("Unable to complete post request")
		log.Fatalln(err)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	log.Println(resp.StatusCode)
}
func GetRecipeByID(id int) Recipe {
	stringid := strconv.Itoa(id)
	url := "http://localhost:8080/recipe?id=" + stringid
	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var recipe Recipe
	err = json.NewDecoder(resp.Body).Decode(&recipe)
	if err != nil {
		log.Fatalln("error: unable to decode response body: ", err)
	}
	return recipe
}
func GetLikedRecipes(u User) []Rating {
	userid := strconv.Itoa(u.User_id)
	url := "http://localhost:8080/getrating/" + userid
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var ratings []Rating
	err = json.NewDecoder(resp.Body).Decode(&ratings)
	if err != nil {
		log.Fatalln("error decoding http response body: ", err)
	}
	return ratings
}
func DeleteRating(ratingID int) {
	stringratingID := strconv.Itoa(ratingID)
	url := "http://localhost:8080/deleterating/" + stringratingID
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Error deleting entry:", resp.Status)
		return
	}
}

func CreateRecipe(recipe Recipe) {
	fmt.Println(recipe)
	url := "http://localhost:8080/create"
	jsonStr, err := json.Marshal(recipe)
	if err != nil {
		log.Println("error marshalling user")
		log.Fatalln(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	log.Println(resp.StatusCode)
}
