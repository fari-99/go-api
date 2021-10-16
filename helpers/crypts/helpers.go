package crypts

import "crypto/rand"

func GenerateRandString(strSize int, randType string) string {
	var dictionary string
	switch randType {
	case "alphanum":
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "alpha":
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "number":
		dictionary = "0123456789"
	default:
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	randString := make([]byte, strSize)
	_, _ = rand.Read(randString)

	for k, v := range randString {
		randString[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(randString)
}
