package routes

import (
	"go-api/configs"
	"go-api/controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupCustomerRoute() *iris.Application {
	log.Println("Setup Customer router")

	app := routes.irisApp
	db := routes.DB
	redis := routes.Redis

	authentication := configs.NewMiddleware(configs.MiddlewareConfiguration{})

	// Approver Endpoint collection
	app.PartyFunc("/customers", func(customers iris.Party) {
		customerController := &controllers.CustomerController{DB: db, Redis: redis}
		//companyIDPathName := "companyID"

		// authentication data
		customers.Post("/auth", customerController.AuthenticateAction)
		customers.Post("/test-redis", customerController.TestRedisAction)

		customers.Post("/", authentication, customerController.CreateAction) // Create
		//customers.Get("/{"+companyIDPathName+":int64}", customerController.ReadAction)    // Read
		//customers.Put("/{"+companyIDPathName+":int64}", customerController.UpdateAction)    // Update
		//customers.Delete("/{"+companyIDPathName+":int64}", customerController.DeleteAction) // Delete

		//customers.PartyFunc("/"+companyIDPathName+"/customer-cars", func(customerCars router.Party) {
		//	customerCarsController := &controllers.CustomerCarController{DB: db}
		//	companyCarIDPathName := "companyCarID"
		//
		//	customerCars.Use(authentication)
		//
		//	customerCars.Post("/", customerCarsController.CreateAction)                                   // Create
		//	customerCars.Get("/{"+companyCarIDPathName+":int64}", customerCarsController.ReadAction)      // Read
		//	customerCars.Put("/{"+companyCarIDPathName+":int64}", customerCarsController.UpdateAction)    // Update
		//	customerCars.Delete("/{"+companyCarIDPathName+":int64}", customerCarsController.DeleteAction) // Delete
		//})
	})

	return app
}
