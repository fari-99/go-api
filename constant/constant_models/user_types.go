package constant_models

import "go-api/constant"

func GetUserTypes() map[int]string {
	userTypes := make(map[int]string)

	userTypes[constant.UserTypeCustomer] = "Customer"
	userTypes[constant.UserTypeSeller] = "Seller"

	return userTypes
}

func GetUserCustomerRoles() map[int]string {
	userRoles := make(map[int]string)

	userRoles[constant.CustomerRolesAdmin] = "Admin"
	userRoles[constant.CustomerRolesPic] = "PIC"
	userRoles[constant.CustomerRolesFinance] = "Finance"
	userRoles[constant.CustomerRolesStaff] = "Staff"

	return userRoles
}

func GetUserSellerRoles() map[int]string {
	userRoles := make(map[int]string)

	userRoles[constant.SellerRolesAdmin] = "Admin"
	userRoles[constant.SellerRolesPic] = "PIC"
	userRoles[constant.SellerRolesFinance] = "Finance"
	userRoles[constant.SellerRolesStaff] = "Staff"

	return userRoles
}
