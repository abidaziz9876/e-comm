package routes

import (
	"github.com/abidaziz9876/e-comm/controllers"
	"github.com/abidaziz9876/e-comm/database"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	client := database.Client
	ProdCollection := client.Database("Ecommerce").Collection("Products")
	UserCollection := client.Database("Ecommerce").Collection("Users")
	incomingRoutes.POST("/users/signup", controllers.SignUp(UserCollection))
	incomingRoutes.POST("/users/login", controllers.Login(UserCollection))
	incomingRoutes.POST("/admin/addproduct", controllers.ProductViewerAdmin(ProdCollection))
	incomingRoutes.GET("/users/productview", controllers.GetAllProducts(ProdCollection))
	incomingRoutes.GET("/users/search", controllers.SearchProductByQuery(ProdCollection))
}
