package routes

import (
	"goService/controllers"
	"goService/utils"
	"log"

	"github.com/kataras/iris"
)

func init() {
	log.Println("Initialize User router")
	app := utils.GetIrisApplication()
	db, _ := utils.DatabaseBase().SetConnection()

	authentication := utils.NewMiddleware(utils.MiddlewareConfiguration{})

	// Approver Endpoint collection
	app.PartyFunc("/customers", func(customers iris.Party) {
		customerController := &controllers.CustomerController{DB: db}
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
}
