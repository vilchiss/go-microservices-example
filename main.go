// Recipes API
//
// This is a sample recipes API
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Luis Oropeza <luis.oropeza@gmail.com> https://www.gitlab.com/luisfi
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"context"
	"go-microservices-example/handlers"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	ctx            context.Context
	err            error
	client         *mongo.Client
	collection     *mongo.Collection
	redisClient    *redis.Client
	recipesHandler *handlers.RecipesHandler
	authHandler    *handlers.AuthHandler
)

func init() {
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	log.Println(status)

	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	authHandler = &handlers.AuthHandler{}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &handlers.Claims{}

		token, err := jwt.ParseWithClaims(tokenValue, claims, func(tkn *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if token == nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Next()
	}
}

func main() {
	router := gin.Default()
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.POST("/signin", authHandler.SignInHandler)
	authorized := router.Group("/")
	authorized.Use(AuthMiddleware())
	authorized.POST("/recipes", recipesHandler.CreateRecipeHandler)
	authorized.GET("/recipes/:id", recipesHandler.GetRecipeByIDHandler)
	authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.Run()
}
