package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	_ "github.com/abidaziz9876/e-comm/docs"
	"github.com/abidaziz9876/e-comm/controllers"
	"github.com/abidaziz9876/e-comm/database"
	"github.com/abidaziz9876/e-comm/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {

	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %s", err)
	}
	database.DBSet()
	client := database.Client
	ProdCollection := client.Database("Ecommerce").Collection("Products")
	UserCollection := client.Database("Ecommerce").Collection("Users")
	fmt.Println(ProdCollection)
	app := controllers.NewApplication(ProdCollection, UserCollection)

	router := gin.Default()
	routes.UserRoutes(router)
	// router.Use(middleware.Authentication())
	router.GET("/addtocart", controllers.AddToCart(ProdCollection, UserCollection))
	router.GET("/removeitem", controllers.RemoveItem(ProdCollection, UserCollection))
	router.GET("/listcart", controllers.GetItemFromCart(UserCollection))
	router.POST("/addaddress", controllers.AddAddress(UserCollection))
	router.PUT("/edithomeaddress", controllers.EditHomeAddress(UserCollection))
	router.PUT("/editworkaddress", controllers.EditWorkAddress(UserCollection))
	router.GET("/deleteaddresses", controllers.DeleteAddress(UserCollection))
	router.GET("/cartcheckout", app.ByFromCart())
	router.GET("/instantbuy", app.InstantBuyer())
	log.Print("Server listening on http://localhost:8085/")
	// url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	router.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := http.ListenAndServe(":8085", router); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

}
