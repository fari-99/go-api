package storages

import (
	"fmt"
	"strconv"
	"strings"
)

func AvailableImageSize(width int64, height int64) (isAvailable bool) {
	availableImageSize := map[string]bool{
		"40x40":    true,
		"54x54":    true,
		"62x62":    true,
		"77x77":    true,
		"119x119":  true,
		"126x127":  true,
		"130x130":  true,
		"174x174":  true,
		"180x180":  true,
		"239x239":  true,
		"343x343":  true,
		"361x203":  true,
		"370x277":  true,
		"450x450":  true,
		"500x290":  true,
		"500x500":  true,
		"600x600":  true,
		"738x415":  true,
		"800x800":  true,
		"900x900":  true,
		"1170x400": true,
		"2000x334": true,
		"1200x200": true,
		"1200x300": true,
	}

	imageSize := fmt.Sprintf("%dx%d", width, height)
	return availableImageSize[imageSize]
}

func GetImageDimensions(imageSize string) (width int64, height int64, isValid bool, err error) {
	imageSizes := strings.Split(imageSize, "x")
	if len(imageSizes) > 2 || len(imageSizes) == 0 {
		return 0, 0, false, fmt.Errorf("image size is not defined")
	}

	width, err = strconv.ParseInt(imageSizes[0], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("width is not integer")
	}

	height, err = strconv.ParseInt(imageSizes[1], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("height is not integer")
	}

	isValid = AvailableImageSize(width, height)
	return
}
