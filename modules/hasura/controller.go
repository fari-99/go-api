package hasura

import (
	"encoding/json"
	"go-api/helpers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service Service
}

func (c controller) UpdateArticle(ctx *gin.Context) {
	var input map[string]interface{}
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	inputMarshal, _ := json.Marshal(input)
	log.Printf(string(inputMarshal))

	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
