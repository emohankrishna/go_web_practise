package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	LastName  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func CreatePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(person)
	fmt.Println(client)
	collection := client.Database("mohandb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	result, err := collection.InsertOne(ctx, person)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		fmt.Println("Failed at result encoder")
	}
}

func CreatePersons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var people []interface{}
	err := json.NewDecoder(r.Body).Decode(&people)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(people)
	collection := client.Database("mohandb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	result, err := collection.InsertMany(ctx, people)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		fmt.Println("Failed at result encoder")
	}
}

func GetPerson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("mohandb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(person)
}
func GetPeople(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	collection := client.Database("mohandb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cursor, _ := collection.Find(ctx, bson.D{})
	result := []Person{}
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		result = append(result, person)
	}
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(w).Encode(result)
}
func init() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println(client.ListDatabaseNames(ctx, bson.D{}))
	fmt.Println("Successfully connected and pinged.")
}
func main() {

	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePerson).Methods("POST")
	router.HandleFunc("/people", CreatePersons).Methods("POST")
	router.HandleFunc("/people", GetPeople).Methods("GET")
	router.HandleFunc("/person/{id}", GetPerson).Methods("GET")
	fmt.Println("Starting the Application..............")
	http.ListenAndServe(":8080", router)

}
