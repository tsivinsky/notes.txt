package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        primitive.ObjectID `bson:"_id"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type Note struct {
	Id        primitive.ObjectID `bson:"_id"`
	Title     string             `bson:"title"`
	Body      string             `bson:"body"`
	Author    primitive.ObjectID `bson:"author"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func main() {
	godotenv.Load()

	mongoUri := os.Getenv("MONGO_URI")

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	db := client.Database("dev")

	r := gin.Default()
	r.LoadHTMLGlob("views/*")
	r.Static("/static", "./public")

	r.GET("/", func(c *gin.Context) {
		userId, err := c.Cookie("user_id")
		if err != nil {
			c.Redirect(302, "/signin")
			return
		}

		userObjectId, _ := primitive.ObjectIDFromHex(userId)

		users := db.Collection("users")

		user := new(User)
		x := users.FindOne(ctx, bson.M{"_id": userObjectId})
		err = x.Decode(&user)
		if err != nil {
			c.Redirect(302, "/signin")
			return
		}

		notes := db.Collection("notes")

		var userNotes []Note
		q, err := notes.Find(ctx, bson.M{"author": user.Id})
		if err == mongo.ErrNoDocuments {
			c.Redirect(302, "/")
			return
		}
		q.All(ctx, &userNotes)

		c.HTML(200, "index.html", gin.H{
			"email": user.Email,
			"notes": userNotes,
		})
	})

	r.GET("/signin", func(c *gin.Context) {
		c.HTML(200, "signin.html", gin.H{})
	})

	r.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", gin.H{})
	})

	r.GET("/logout", func(c *gin.Context) {
		_, err := c.Cookie("user_id")
		if err != nil {
			c.Redirect(302, "/")
			return
		}

		c.SetCookie("user_id", "", int(1*time.Second), "/", "", true, true)

		c.Redirect(302, "/signin")
	})

	r.POST("/signin", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "" || password == "" {
			c.Redirect(302, "/signin")
			return
		}

		users := db.Collection("users")

		var user User

		x := users.FindOne(ctx, bson.M{"email": email})
		err := x.Decode(&user)
		if err != nil {
			fmt.Println(err.Error())
			c.Redirect(302, "/signin")
			return
		}

		// TODO: compare password with hash in database
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			c.Redirect(302, "/signin")
			return
		}

		c.SetCookie("user_id", user.Id.Hex(), int(48*time.Hour), "/", "", true, true)

		c.Redirect(302, "/")

		c.Redirect(302, "/")
	})

	r.POST("/signup", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "" || password == "" {
			c.Redirect(302, "/signin")
			return
		}

		users := db.Collection("users")

		var user User
		x := users.FindOne(ctx, bson.M{"email": email})
		err := x.Decode(&user)
		if err != mongo.ErrNoDocuments {
			c.Redirect(302, "/signup")
			return
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

		user = User{
			Id:        primitive.NewObjectID(),
			Email:     email,
			Password:  string(hash),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		q, err := users.InsertOne(ctx, &user)
		if err != nil {
			fmt.Println(err.Error())
			c.Redirect(302, "/signin")
			return
		}

		id := q.InsertedID.(primitive.ObjectID).Hex()

		c.SetCookie("user_id", id, int(48*time.Hour), "/", "", true, true)

		c.Redirect(302, "/")
	})

	r.POST("/notes", func(c *gin.Context) {
		userId, err := c.Cookie("user_id")
		if err != nil {
			c.Redirect(302, "/signin")
		}
		userObjectId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			fmt.Println(err.Error())
			c.Redirect(302, "/")
			return
		}

		title := c.PostForm("title")
		body := c.PostForm("body")

		notes := db.Collection("notes")

		note := Note{
			Id:        primitive.NewObjectID(),
			Title:     title,
			Body:      body,
			Author:    userObjectId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		notes.InsertOne(ctx, &note)

		c.Redirect(302, "/")
	})

	port := getPortAddr(":5000")
	r.Run(port)
}

func getPortAddr(fallbackPort string) string {
	port := os.Getenv("PORT")

	if port == "" {
		port = fallbackPort
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return port
}
