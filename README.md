# Welcome to Frigid

## What is Frigid?

Frigid is a CLI application that allows users to find and save recipes using ingredients they have on hand.
Users can also upload their own recipes which can then be accessed by other users.

## How was it made?

Frigid is written entirely in Go and uses the gorilla/mux package to host a REST API that interacts with a postgres database. The recipes were seeded using the Tasty API offered on RapidAPI.
Users interact with a CLI interface that then sends http requests to that API and executes the appropriate request. Besides from the standard go library and gorilla/mux,
this program also uses the godotenv package to obscure API keys and the pq package to communicate with the Postgres DB.

