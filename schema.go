package main

//schema and resolvers are here
import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/graphql-go/graphql"
)

// Helper function to import json from file to map
// then we can set out data from json file to BeastList and use as source of truth
func importJSONDataFromFile(fileName string, result interface{}) (isOK bool) {
	isOK = true
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Print("Error:", err)
		isOK = false
	}
	err = json.Unmarshal(content, result)
	if err != nil {
		isOK = false
		fmt.Print("Error:", err)
	}
	return
}

var BeastList []Beast

// setting to a BeastList Variable address of memory
var _ = importJSONDataFromFile("./beastData.json", &BeastList)

// camelcase json -> pascalcase golang
// using json tags to map fields to beastType
type Beast struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OtherNames  []string `json:"otherNames"`
	ImageURL    string   `json:"imageUrl"`
}

// define custom GraphQL ObjectType `beastType` for our Golang struct `Beast`
// Note that
// - the fields in our todoType maps with the json tags for the fields in our struct
// - the field type matches the field type in our struct
// check root query for implementation
var beastType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Beast",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"otherNames": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"imageUrl": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var currentMaxId = 5

// root mutation
// this should be just like rootquery with differences in return
var rootMutation = graphql.NewObject(graphql.ObjectConfig{ //adding beast to list
	Name: "RootMutation",
	Fields: graphql.Fields{
		"addBeast": &graphql.Field{
			Type:        beastType, // the return type for this field
			Description: "add a new beast",
			Args: graphql.FieldConfigArgument{ //we want to query beast types by name
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String), //query by name so use string
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"otherNames": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
				"imageUrl": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},

			//writing the resolvers
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				// marshall and cast the argument value
				name, _ := params.Args["name"].(string)
				description, _ := params.Args["description"].(string)
				otherNames, _ := params.Args["otherNames"].([]string)
				imageUrl, _ := params.Args["imageUrl"].(string)

				// figure out new id
				newID := currentMaxId + 1
				currentMaxId = currentMaxId + 1

				// perform mutation operation here
				// for e.g. create a Beast and save to DB.
				newBeast := Beast{
					ID:          newID,
					Name:        name,
					Description: description,
					OtherNames:  otherNames,
					ImageURL:    imageUrl,
				}

				BeastList = append(BeastList, newBeast)

				// return the new Beast object that we supposedly save to DB
				// Note here that
				// - we are returning a `Beast` struct instance here
				// - we previously specified the return Type to be `beastType`
				// - `Beast` struct maps to `beastType`, as defined in `beastType` ObjectConfig`
				return newBeast, nil
			},
		},
		"updateBeast": &graphql.Field{
			Type:        beastType, // the return type for this field
			Description: "Update existing beast",
			//We want to query beast by name of type string
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"otherNames": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
				"imageUrl": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["id"].(int)
				affectedBeast := Beast{}

				// Search list for beast with id
				for i := 0; i < len(BeastList); i++ {
					if BeastList[i].ID == id {
						if _, ok := params.Args["description"]; ok {
							BeastList[i].Description = params.Args["description"].(string)
						}
						if _, ok := params.Args["name"]; ok {
							BeastList[i].Name = params.Args["name"].(string)
						}
						if _, ok := params.Args["imageUrl"]; ok {
							BeastList[i].ImageURL = params.Args["imageUrl"].(string)
						}
						if _, ok := params.Args["otherNames"]; ok {
							BeastList[i].OtherNames = params.Args["otherNames"].([]string)
						}
						// reassigning updated beast so we can return it
						// breaks out of loop so we can return after everything is done
						affectedBeast = BeastList[i]
						break
					}
				}
				// Return affected beast, new list with beasts that were updated
				return affectedBeast, nil
			},
		},
	},
})

// root query, setting up the schema
// test with Sandbox at localhost:8080/sandbox
var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery", //object Query Type
	Fields: graphql.Fields{
		//first field
		"beast": &graphql.Field{ //graphql.Fiel makes a map to type,des,args,resolver
			Type:        beastType,
			Description: "Get single beast",
			//We want to query beast by name of type string
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			//Resolver
			//looks through array of beasts
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				nameQuery, isOK := params.Args["name"].(string)
				if isOK {
					// Search for el with name
					for _, beast := range BeastList {
						if beast.Name == nameQuery {
							return beast, nil
						}
					}
				}

				return Beast{}, nil
			},
		},

		"beastList": &graphql.Field{ //second field
			Type:        graphql.NewList(beastType),
			Description: "List of beasts",

			//Resolver
			//just returns a list of beasts
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return BeastList, nil
			},
		},
	},
})

// define schema, with our rootQuery and rootMutation
// gets server up and running
var BeastSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    rootQuery,
	Mutation: rootMutation, //adding beast to the list
})
