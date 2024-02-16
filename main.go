package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name"`
	Email string             `json:"email"`
	Role  string             `json:"role"`
	Avt   string             `json:"avt"`
}

type Transaction struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId     primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	CategoryId primitive.ObjectID `json:"categoryId,omitempty" bson:"categoryId,omitempty"`
	TypeId     primitive.ObjectID `json:"typeId,omitempty" bson:"typeId,omitempty"`
	Amount     int                `json:"amount" bson:"amount"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

type Category struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name"`
	Label string             `json:"label"`
}

type Type struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name"`
	Label string             `json:"label"`
}

type Card struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId   primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	BankId   primitive.ObjectID `json:"bankId,omitempty" bson:"bankId,omitempty"`
	ExpireIn time.Time          `json:"expireIn"`
	Number   string             `json:"number"`
}

type Bank struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name"`
}

var client *mongo.Client

var secretKey = "76766d67364e35119f0c5a198a173fb980452ae08347a79c8e419239425bebcd"

func getUserIdStr(r *http.Request, c *gin.Context) string {
	//Get user id from token
	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No bearer token provided"})
		return ""
	}
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	fmt.Println("Token: ", reqToken)
	userIdStr := validateToken(reqToken, c)

	return userIdStr
}

func connectDB() {
	connectionString := "mongodb+srv://doantranbadat:NJpsmGtvzUuC0LLG@cluster0.6qe8hwu.mongodb.net"
	clientOptions := options.Client().ApplyURI(connectionString)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	fmt.Println("Client: ", client)
}

func getAllUsers(c *gin.Context) {
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB client is not initialized"})
		return
	}

	collection := client.Database("Dapex").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer cur.Close(ctx)

	var users []User
	for cur.Next(ctx) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func getAllCategory(c *gin.Context) {
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB client is not initialized"})
		return
	}

	collection := client.Database("Dapex").Collection("categories")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
		return
	}
	defer cur.Close(ctx)

	var categories []Category
	for cur.Next(ctx) {
		var category Category
		err := cur.Decode(&category)
		if err != nil {
			log.Println(err)
			continue
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, categories)
}

func getAllType(c *gin.Context) {
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB client is not initialized"})
		return
	}

	collection := client.Database("Dapex").Collection("types")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
		return
	}
	defer cur.Close(ctx)

	var types []Type
	for cur.Next(ctx) {
		var typeItem Type
		err := cur.Decode(&typeItem)
		if err != nil {
			log.Println(err)
			continue
		}
		types = append(types, typeItem)
	}

	c.JSON(http.StatusOK, types)
}

func addNewTransaction(c *gin.Context) {
	var newTransaction Transaction
	if err := c.BindJSON(&newTransaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if newTransaction.TypeId.IsZero() || newTransaction.CategoryId.IsZero() || newTransaction.UserId.IsZero() || newTransaction.Amount == 0 {
		fmt.Println("Missing transaction information")
	}

	newTransaction.CreatedAt = time.Now()

	collection := client.Database("Dapex").Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newTransaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, newTransaction)
}

func deleteTransaction(r *http.Request, c *gin.Context) {

	var transactionDel Transaction

	if err := c.BindJSON(&transactionDel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	filter := bson.M{
		"_id":        transactionDel.ID,
		"userId":     transactionDel.UserId,
		"categoryId": transactionDel.CategoryId,
		"typeId":     transactionDel.TypeId,
		"amount":     transactionDel.Amount,
	}

	userIdStr := getUserIdStr(r, c)

	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		// Xử lý lỗi khi chuyển đổi không thành công
		fmt.Println("Error converting userIDString to ObjectID:", err)
		return
	}

	if userId != transactionDel.UserId {
		c.JSON(http.StatusBadRequest, "Unauthorized")
		return
	}

	collection := client.Database("Dapex").Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Printf("Error deleting transaction: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, "Delete successfully")
	fmt.Println(result)
}

func updateTransaction(c *gin.Context) {
	var transactionUpdate Transaction

	if err := c.BindJSON(&transactionUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	filter := bson.M{"_id": transactionUpdate.ID}

	updateData := bson.M{
		"$set": bson.M{
			"amount": transactionUpdate.Amount,
		},
	}

	collection := client.Database("Dapex").Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		fmt.Printf("Error deleting transaction: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, transactionUpdate)
}

func getUserTransaction(r *http.Request, c *gin.Context) {

	userIdStr := getUserIdStr(r, c)

	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		// Xử lý lỗi khi chuyển đổi không thành công
		fmt.Println("Error converting userIDString to ObjectID:", err)
		return
	}

	//Get current user transaction
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB client is not initialized"})
		return
	}

	collection := client.Database("Dapex").Collection("transactions")

	filter := bson.M{"userId": userId}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
		return
	}
	defer cur.Close(ctx)

	var userTransactions []Transaction
	for cur.Next(ctx) {
		var userTransaction Transaction
		err := cur.Decode(&userTransaction)
		if err != nil {
			log.Println(err)
			continue
		}
		userTransactions = append(userTransactions, userTransaction)
	}
	c.JSON(http.StatusOK, userTransactions)
}

func validateToken(accessToken string, c *gin.Context) string {
	if accessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No bearer token provided"})
		return ""
	}

	// Xác thực token và lấy claims
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		fmt.Println("Error when parse token")
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		fmt.Println("Invalid token")
		return ""
	}
	fmt.Println("Claims: ", claims)

	userID, ok := claims["userId"].(string)
	if !ok {
		fmt.Println("Not ok!")
		return ""
	}

	return userID
}

func main() {
	connectDB()

	router := gin.Default()

	// Use CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Replace with your allowed origins
	config.AllowMethods = []string{"GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH"}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}
	router.Use(cors.New(config))

	// Define a route to get all users
	router.GET("/api/v1/users", getAllUsers)
	router.GET("/api/v1/categories", getAllCategory)
	router.GET("/api/v1/types", getAllType)
	router.POST("/api/v1/transaction", addNewTransaction)
	router.PATCH("/api/v1/transaction", updateTransaction)
	router.POST("/api/v1/transaction/delete", func(c *gin.Context) {
		deleteTransaction(c.Request, c)
	})
	router.GET("api/v1/user/transactions", func(c *gin.Context) {
		getUserTransaction(c.Request, c)
	})

	// Run the server on port 4000
	port := "4000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("Server is running on port 111 %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
