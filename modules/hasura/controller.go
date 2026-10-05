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

// UpdateArticle godoc
// @Summary      Hasura article webhook
// @Tags         hasura
// @Accept       json
// @Produce      json
// @Param        body  body  object  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /hasura/events/articles [put]
// @Router       /hasura/crons/articles [post]
// @Router       /hasura/schedule/articles [post]
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
