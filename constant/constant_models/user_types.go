package constant_models

import "go-api/constant"

func GetUserTypes() map[int]string {
	userTypes := make(map[int]string)

	userTypes[constant.UserTypeCustomer] = "Customer"
	userTypes[constant.UserTypeSeller] = "Seller"

	return userTypes
}