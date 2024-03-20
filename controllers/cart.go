package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/abidaziz9876/e-comm/database"
	"github.com/abidaziz9876/e-comm/models"
	"github.com/abidaziz9876/e-comm/responses"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}




// AddToCartTags				godoc
// @Tags 						Cart Apis
// @Summary    					to add in cart
// @Description 				It will just add the product in the user cart
// @Param						id query string true "id"
// @Param						userID query string true "userID"
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/addtocart [GET]
func AddToCart(prodCollection, userCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {

		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProducToCart(ctx, prodCollection, userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    nil,
			Message: "Successfully added to the cart",
		})
	}
}



// AddToCartTags				godoc
// @Tags 						Cart Apis
// @Summary    					to remove in cart
// @Description 				It will just remove the product from the user cart
// @Param						id query string true "id"
// @Param						userID query string true "userID"
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/removeitem [GET]
func RemoveItem(prodCollection, userCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {


		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is inavalid")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		ProductID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.RemoveCartItem(ctx, prodCollection, userCollection, ProductID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    nil,
			Message: "Successfully removed from the cart",
		})
	}
}




func GetItemFromCart(UserCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}

		usert_id, _ := primitive.ObjectIDFromHex(user_id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledcart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usert_id}}).Decode(&filledcart)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(500, "not id found")
			return
		}

		filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usert_id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		if err != nil {
			log.Println(err)
		}
		var listing []bson.M
		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledcart.UserCart)
		}
		ctx.Done()
	}
}


// AddToCartTags				godoc
// @Tags 						Cart Apis
// @Summary    					to buy from cart
// @Description 				It will just buy the product from the user cart
// @Param						id query string true "id"
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/cartcheckout [GET]
func (app *Application) ByFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    nil,
			Message: "Successfully placed the order",
		})
	}
}



// AddToCartTags				godoc
// @Tags 						Cart Apis
// @Summary    					to buy instant
// @Description 				It will just buy the product instantly
// @Param						userid query string true "userid"
// @Param						pid query string true "pid"
// @Produce						application/json
// @Success						200 {object} responses.ApplicationResponse{}
// @Router						/instantbuy [GET]
func (app *Application) InstantBuyer() gin.HandlerFunc {
	return func(c *gin.Context) {
		UserQueryID := c.Query("userid")
		if UserQueryID == "" {
			log.Println("UserID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}
		ProductQueryID := c.Query("pid")
		if ProductQueryID == "" {
			log.Println("Product_ID id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product_id is empty"))
		}
		productID, err := primitive.ObjectIDFromHex(ProductQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, productID, UserQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  200,
			Data:    nil,
			Message: "Successfully placed the order",
		})
	}
}
