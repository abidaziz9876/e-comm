package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"


	"github.com/abidaziz9876/e-comm/models"
	"github.com/abidaziz9876/e-comm/responses"
	"github.com/abidaziz9876/e-comm/tokens"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
// var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")

var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login Or Passowrd is Incorerct"
		valid = false
	}
	return valid, msg
}

// SignupTags					godoc
// @Tags 						User Apis
// @Summary    					User SignUp
// Accept						json
// @Param user body models.User true "User"
// @Description 				Please provide firstname, lastname, phone, email and password to signup
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/users/signup [POST]
func SignUp(UserCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}
		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, _ := tokens.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    user,
			Message: "Successfully SignUp",
		})

	}
}


// SignInTags					godoc
// @Tags 						User Apis
// @Summary    					User SignIn
// Accept						json
// @Param user body models.User true "User"
// @Description 				Please provide email and password to signin
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/users/login [POST]
func Login(UserCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}
		PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		defer cancel()
		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := tokens.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		defer cancel()
		tokens.UpdateAllTokens(token, refreshToken, founduser.User_ID, UserCollection)
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    founduser,
			Message: "Successfully Logged In",
		})

	}
}





// ProductAddTags				godoc
// @Tags 						User Apis
// @Summary    					To add a product in the store
// Accept						json
// @Param product body models.Product true "Product"
// @Description 				you can add products here by filling product details
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/admin/addproduct [POST]
func ProductViewerAdmin(ProductCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.Product_ID = primitive.NewObjectID()
		_, anyerr := ProductCollection.InsertOne(ctx, products)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    products,
			Message: "Successfully Added a Product",
		})
	}
}


// AllProductTags				godoc
// @Tags 						User Apis
// @Summary    					To show all products
// @Description 				It will give you all the products
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/users/productview [GET]
func GetAllProducts(ProductCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var productlist []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Someting Went Wrong Please Try After Some Time")
			return
		}
		err = cursor.All(ctx, &productlist)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			// Don't forget to log errors. I log them really simple here just
			// to get the point across.
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    productlist,
			Message: "Successfully fetched all Products",
		})

	}
}



// GetProductByQuery			godoc
// @Tags 						User Apis
// @Summary    					To get product
// @Description 				It will give you the product for product id
// @Param						name query string true "name"
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/users/search [GET]
func SearchProductByQuery(ProductCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchproducts []models.Product
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			c.Abort()
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		searchquerydb, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
		if err != nil {
			c.IndentedJSON(404, "something went wrong in fetching the dbquery")
			return
		}
		err = searchquerydb.All(ctx, &searchproducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer searchquerydb.Close(ctx)
		if err := searchquerydb.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid request")
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    searchproducts,
			Message: "Successfully fetched the product",
		})
	}
}
