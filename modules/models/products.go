package models

import "time"

// a building have many rooms,
// the rooms will be used as products
// each product can have seats (seats and no seats)
// each product can have seats categories
// every seat in a room have their own configuration depends on their categories

type Buildings struct {
	Base
	Name         string // Balai Sarbini
	Descriptions string // Balai Sarbini adalah .....
	Address      string // Jakarta
	Longitude    string
	Latitude     string
	Status       int8 // active, not active
	Rooms        []Rooms
}

type Rooms struct {
	Base
	BuildingID   string
	Name         string // Balai Sarbini
	Descriptions string // Stadium
	Status       int8   // active, not active
	TotalSeats   int64  // 1500 seats
	Products     []Products
}

type ProductType struct {
	Base
	Name         string // Seats, Entry
	Descriptions string
	Status       int8
}

type Products struct {
	Base
	BuildingID    string
	RoomID        string
	ProductTypeID int8
	Name          string // VVIP, PLATINUM, GOLD, SILVER, BRONZE
	AreaCode      string
	Price         float64 // 3jt, 2jt, 1jt, 750rb, 500rb
	TotalSeat     int64   // 50, 50, 100, 150, 200, 300
	Status        int8    // active, not active

	// EventDateStart
	// date for event start (usually no more buying)
	EventDateStart time.Time
	// EventDateEnd
	// date for event end
	EventDateEnd time.Time
	// BookDateStart
	// date to buy the product start
	BookDateStart time.Time
	// BookDateEnd
	// date to buy the product end (if this overlap with EventDateStart or EventDateEnd,
	// then IsBuyOnEvent must be true
	BookDateEnd time.Time

	// CanBuyOnEvent
	// flags to check if you can still buy product when event already starting
	CanBuyOnEvent int8

	ProductType        ProductType
	SeatConfigurations []SeatConfigurations
}

type SeatConfigurations struct {
	Base
	ProductID       string
	PhysicalRowName string
	Status          int8
	Seats           []Seats
}

type Seats struct {
	Base
	SeatConfigurationID string
	Name                string
	SeatNumber          int  // 1,2,3
	GridNumber          int  // 1,2,3
	Status              int8 // free, reserved, paid
}
