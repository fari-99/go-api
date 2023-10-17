package flip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fari-99/go-flip"
	"github.com/fari-99/go-flip/constants"
	"github.com/fari-99/go-flip/models"
	"github.com/fari-99/go-flip/requests"
	"github.com/fari-99/go-helper/storages"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cast"

	"go-api/helpers"
	"go-api/modules/configs"
)

type controller struct {
	*configs.DI
}

type ErrorResponse struct {
	Message        string      `json:"message"`
	ErrorMessage   string      `json:"error_message"`
	ErrorDetails   interface{} `json:"error_details,omitempty"`
	IdempotencyKey string      `json:"idempotency_key,omitempty"`
	RequestID      string      `json:"request_id,omitempty"`
}

func getFiles(ctx *gin.Context, fileName, paramName string) (uploadFileData *requests.UploadFile, err error) {
	fileHeader, err := ctx.FormFile(paramName)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, err
	}

	defer file.Close()

	buffer := make([]byte, 1024)
	_, _ = file.Seek(0, 0)
	_, err = file.Read(buffer)
	if err != nil {
		err = fmt.Errorf("file could not be read, err := %s", err.Error())
		return nil, err
	}

	_, _ = file.Seek(0, 0)
	contentType := http.DetectContentType(buffer)

	var isImage = true

	switch contentType {
	case storages.ContentTypePNG, storages.ContentTypeJPEG, storages.ContentTypeJPG:
		isImage = true
	default:
		isImage = false
	}

	if !isImage {
		return nil, fmt.Errorf("file is not image or file type invalid (not jpg, png, or jpeg")
	}

	uploadFileData = &requests.UploadFile{
		ContentType: contentType,
		FileName:    fileName,
		File:        buf,
		FileByte:    buf.Bytes(),
	}

	return uploadFileData, nil
}

func (c controller) IsMaintenance(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	isMaintenance, err := baseFlip.IsMaintenance()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to check flip is maintenance or not",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})

		return
	}

	helpers.NewResponse(ctx, http.StatusOK, isMaintenance)
	return
}

func (c controller) GetBalance(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	balanceModel, err := baseFlip.GetBalance()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get account balance",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, balanceModel)
	return
}

func (c controller) GetBankInfo(ctx *gin.Context) {
	var input models.GetBankInfoRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	baseFlip := flip.NewBaseFlip()
	bankList, err := baseFlip.GetBankInfo(input)

	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get bank info",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bankList)
	return
}

func (c controller) BankInquiry(ctx *gin.Context) {
	var input models.SendBankAccountInquiryRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	baseFlip := flip.NewBaseFlip()
	inquiryModel, err := baseFlip.SendBankAccountInquiry(input)

	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to send bank account inquiry",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, inquiryModel)
	return
}

// --------------------------------------------------------------

func (c controller) CreateDisbursement(ctx *gin.Context) {
	var input models.CreateDisbursementRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(uuid.New().String())
	inquiryModel, idempotencyKey, err := baseFlip.CreateDisbursement(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create disbursement",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idempotencyKey,
		})
		return
	}

	inquiryModel.IdempotencyKey = idempotencyKey
	helpers.NewResponse(ctx, http.StatusOK, inquiryModel)
	return
}

func (c controller) GetAllDisbursement(ctx *gin.Context) {
	var input models.GetAllDisbursementRequest
	input.Pagination = ctx.DefaultQuery("pagination", "10")
	input.Page = ctx.DefaultQuery("page", "1")
	input.Sort = ctx.DefaultQuery("sort", "id")
	attribute := ctx.DefaultQuery("attribute", "")
	input.Attribute = &attribute

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(uuid.New().String())
	inquiryModel, idempotencyKey, err := baseFlip.GetAllDisbursement(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to get all disbursement",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idempotencyKey,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, inquiryModel)
	return
}

func (c controller) GetDetailDisbursement(ctx *gin.Context) {
	idempotencyKey := ctx.DefaultQuery("idempotency_key", "")
	id := ctx.DefaultQuery("id", "")

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(uuid.New().String())

	var disbursementModel *models.DisbursementModel
	var err error
	if idempotencyKey != "" {
		disbursementModel, err = baseFlip.GetDisbursementByIdemKey(idempotencyKey)
	} else if id != "" {
		disbursementModel, err = baseFlip.GetDisbursementById(cast.ToInt64(id))
	} else {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message:        "get details using idempotency key or disbursement id",
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idempotencyKey,
		})
		return
	}

	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to get disbursement details",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idempotencyKey,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, disbursementModel)
	return
}

// --------------------------------------------------------------

func (c controller) CreateSpecialDisbursement(ctx *gin.Context) {
	var input models.CreateSpecialDisbursementRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	idempotencyKey := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(idempotencyKey)
	specialDisbursement, idemKey, err := baseFlip.CreateSpecialDisbursement(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create special money transfer disbursement",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idemKey,
		})
		return
	}

	specialDisbursement.IdempotencyKey = idempotencyKey
	helpers.NewResponse(ctx, http.StatusOK, specialDisbursement)
	return
}

func (c controller) GetDisbursementCountryList(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	countryList, err := baseFlip.GetDisbursementCountryList()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get country list for disbursement",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, countryList)
	return
}

func (c controller) GetDisbursementCityList(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	countryList, err := baseFlip.GetDisbursementCityList()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get city list for disbursement",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, countryList)
	return
}

func (c controller) GetDisbursementCountyCityList(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	countryList, err := baseFlip.GetDisbursementCountryCityList()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get country-city list for disbursement",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, countryList)
	return
}

// --------------------------------------------------------------

func (c controller) CreateAgentDisbursement(ctx *gin.Context) {
	var input models.CreateDisbursementAgentRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	idempotencyKey := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(idempotencyKey)
	disbursement, idemKey, err := baseFlip.CreateDisbursementAgent(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create agent money transfer disbursement",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idemKey,
		})
		return
	}

	disbursement.IdempotencyKey = idempotencyKey
	helpers.NewResponse(ctx, http.StatusOK, disbursement)
	return
}

func (c controller) ListAgentDisbursement(ctx *gin.Context) {
	var input models.GetDisbursementAgentListRequest
	input.AgentId = ctx.DefaultQuery("agent_id", "0")
	input.Pagination = ctx.DefaultQuery("pagination", "10")
	input.Page = ctx.DefaultQuery("page", "1")
	input.Sort = ctx.DefaultQuery("sort", "desc")

	baseFlip := flip.NewBaseFlip()
	disbursementModels, err := baseFlip.GetDisbursementAgentList(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get list agent disbursement",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, disbursementModels)
	return
}

func (c controller) DetailAgentDisbursement(ctx *gin.Context) {
	transactionIDParam := ctx.Param("transactionID")
	transactionID := cast.ToInt64(transactionIDParam)
	if transactionID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "transaction ID is invalid (less and equal to 0)",
		})
		return
	}

	baseFlip := flip.NewBaseFlip()
	disbursementModel, err := baseFlip.GetDisbursementAgentByID(transactionID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get agent disbursement by ID",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, disbursementModel)
	return
}

// --------------------------------------------------------------

func (c controller) CreateAgents(ctx *gin.Context) {
	var input models.CreateAgentIdentityRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	baseFlip := flip.NewBaseFlip()
	agentIdentity, err := baseFlip.CreateAgentIdentity(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to create agent identity",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, agentIdentity)
	return
}

func (c controller) UpdateAgent(ctx *gin.Context) {
	var input models.UpdateAgentIdentityRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	input.Id = agentID

	baseFlip := flip.NewBaseFlip()
	agentIdentity, err := baseFlip.UpdateAgentIdentity(agentID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to update agent identity",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, agentIdentity)
	return
}

func (c controller) GetAgent(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	baseFlip := flip.NewBaseFlip()
	agentIdentity, err := baseFlip.GetAgentIdentity(agentID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get agent identity",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, agentIdentity)
	return
}

func (c controller) UploadAgentImage(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	var input models.UploadAgentIdentityRequest
	input.Selfie = ctx.DefaultPostForm("selfie", "0")
	input.UserType = ctx.DefaultPostForm("user_type", "1")
	input.IdentityType = ctx.DefaultPostForm("identity_type", "1")

	identityType, err := constants.GetIdentityTypeLabel(input.IdentityType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			ErrorMessage: err.Error(),
		})
		return
	}

	updateFileData, err := getFiles(ctx, identityType, "image")
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message:      "image input is invalid",
			ErrorMessage: err.Error(),
		})
		return
	}

	inputFile := models.UploadAgentIdentityFileRequest{
		Image: *updateFileData,
	}

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	imageFlip, reqID, err := baseFlip.UploadAgentIdentityImage(agentID, requestID, input, inputFile)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to upload agent identity image",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, imageFlip)
	return
}

func (c controller) UploadAgentDocuments(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	supportingDocumentNames := map[string]bool{
		"student_card":            true, // optional
		"student_card_selfie":     true, // optional
		"employee_card":           true, // optional
		"employee_card_selfie":    true, // optional
		"last_certificate":        true, // optional
		"last_certificate_selfie": true, // optional
		"passport":                true, // optional
		"passport_selfie":         true, // optional
		"family_card":             true, // optional
		"family_card_selfie":      true, // optional
		"driving_license":         true, // optional
		"driving_license_selfie":  true, // optional
		"married_card":            true, // optional
		"married_card_selfie":     true, // optional
		"npwp":                    true, // optional
		"npwp_selfie":             true, // optional
		"bpjs_kesehatan":          true, // optional
		"bpjs_kesehatan_selfie":   true, // optional
	}

	supportingDocuments := make(map[string]requests.UploadFile)
	for paramName, check := range supportingDocumentNames {
		if !check {
			continue
		}

		uploadFileData, err := getFiles(ctx, paramName, paramName)
		if errors.Is(err, http.ErrMissingFile) {
			continue
		} else if err != nil {
			helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
				Message:      "error getting uploaded files",
				ErrorMessage: err.Error(),
			})
			return
		}

		supportingDocuments[paramName] = *uploadFileData
	}

	if len(supportingDocuments) == 0 { // no files got uploaded
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message: "no files uploaded",
		})
		return
	}

	var input models.UploadSupportingDocumentsRequest
	input.UserType = ctx.DefaultPostForm("user_type", "")
	input.UserId = agentID

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	imageFlip, reqID, err := baseFlip.UploadSupportingDocuments(requestID, input, supportingDocuments)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to upload agent identity documents",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, imageFlip)
	return
}

func (c controller) SubmitAgent(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	var input models.KycSubmissionRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	message, reqID, err := baseFlip.KYCSubmissions(agentID, requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to submit agent",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, message)
	return
}

func (c controller) RepairAgent(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	var input models.RepairDataRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	repairedData, reqID, err := baseFlip.RepairData(agentID, requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to repaid agent data",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, repairedData)
	return
}

func (c controller) RepairImage(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	var input models.RepairIdentityImageRequest
	input.UserType = ctx.DefaultPostForm("user_type", "")

	uploadFileData, err := getFiles(ctx, "ktp/passport", "image")
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message:      "image input is invalid",
			ErrorMessage: err.Error(),
		})
		return
	}

	inputFile := models.RepairIdentityImageFileRequest{Image: *uploadFileData}
	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	repairedData, reqID, err := baseFlip.RepairIdentityImage(agentID, requestID, input, inputFile)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to repair agent image data",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, repairedData)
	return
}

func (c controller) RepairSelfieImage(ctx *gin.Context) {
	agentIDParam := ctx.Param("agentID")
	agentID := cast.ToInt64(agentIDParam)
	if agentID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "agent ID is invalid (less and equal to 0)",
		})
		return
	}

	var input models.RepairIdentityImageRequest
	input.UserType = ctx.DefaultPostForm("user_type", "")

	uploadFileData, err := getFiles(ctx, "ktp/passport", "image")
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message:      "image input is invalid",
			ErrorMessage: err.Error(),
		})
		return
	}

	inputFile := models.RepairIdentityImageFileRequest{Image: *uploadFileData}
	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	repairedData, reqID, err := baseFlip.RepairIdentitySelfieImage(agentID, requestID, input, inputFile)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to repair agent image selfie data",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, repairedData)
	return
}

func (c controller) AgentCountryList(ctx *gin.Context) {
	var input models.LocationKycRequest
	input.UserType = ctx.DefaultQuery("user_type", "1")

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	locationData, reqID, err := baseFlip.GetCountryList(requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get kyc country list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationData)
	return
}

func (c controller) AgentProvinceList(ctx *gin.Context) {
	var input models.LocationKycRequest
	input.UserType = ctx.DefaultQuery("user_type", "1")
	input.CountryID = cast.ToInt64(ctx.DefaultQuery("country_id", "1"))

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	locationData, reqID, err := baseFlip.GetProvinceList(requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get kyc province list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationData)
	return
}

func (c controller) AgentCityList(ctx *gin.Context) {
	var input models.LocationKycRequest
	input.UserType = ctx.DefaultQuery("user_type", "1")
	input.ProvinceID = cast.ToInt64(ctx.DefaultQuery("province_id", "1"))

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	locationData, reqID, err := baseFlip.GetCityList(requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get kyc city list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationData)
	return
}

func (c controller) AgentDistrictList(ctx *gin.Context) {
	var input models.LocationKycRequest
	input.UserType = ctx.DefaultQuery("user_type", "1")
	input.CityID = cast.ToInt64(ctx.DefaultQuery("city_id", "1"))

	requestID := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	locationData, reqID, err := baseFlip.GetDistrictList(requestID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get kyc district list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
			RequestID:    reqID,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationData)
	return
}

// --------------------------------------------------------------

func (c controller) CreateBill(ctx *gin.Context) {
	var input models.CreateBillRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	idempotencyKey := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(idempotencyKey)
	bill, idemKey, err := baseFlip.CreateBill(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create billing",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idemKey,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) EditBill(ctx *gin.Context) {
	billIDParam := ctx.Param("billID")
	billID := cast.ToInt64(billIDParam)
	if billID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "bill ID (transaction ID) is invalid (less and equal to 0)",
		})
		return
	}

	var input models.EditBillingRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.EditBill(billID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to update billing",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) GetBillDetail(ctx *gin.Context) {
	billIDParam := ctx.Param("billID")
	billID := cast.ToInt64(billIDParam)
	if billID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "bill ID (transaction ID) is invalid (less and equal to 0)",
		})
		return
	}

	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.GetBill(billID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get billing",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) GetBillList(ctx *gin.Context) {
	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.GetAllBill()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get billing list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) GetBillPayments(ctx *gin.Context) {
	billIDParam := ctx.Param("billID")
	billID := cast.ToInt64(billIDParam)
	if billID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "bill ID (transaction ID) is invalid (less and equal to 0)",
		})
		return
	}

	var input models.GetPaymentRequest
	input.Pagination = ctx.DefaultQuery("pagination", "10")
	input.Page = ctx.DefaultQuery("page", "1")
	input.StartDate = ctx.DefaultQuery("start_date", "")
	input.EndDate = ctx.DefaultQuery("end_date", "")
	input.SortBy = ctx.DefaultQuery("sort_by", "id")
	input.SortType = ctx.DefaultQuery("sort_type", "sort_desc")

	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.GetPayment(billID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get billing payments list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) GetPaymentList(ctx *gin.Context) {
	var input models.GetAllPaymentRequest
	input.Pagination = ctx.DefaultQuery("pagination", "10")
	input.Page = ctx.DefaultQuery("page", "1")
	input.StartDate = ctx.DefaultQuery("start_date", "")
	input.EndDate = ctx.DefaultQuery("end_date", "")
	input.SortBy = ctx.DefaultQuery("sort_by", "id")
	input.SortType = ctx.DefaultQuery("sort_type", "sort_desc")

	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.GetAllPayment(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get payments list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

func (c controller) ConfirmBillPayment(ctx *gin.Context) {
	transactionIDParam := ctx.Param("transactionID")
	transactionID := cast.ToInt64(transactionIDParam)
	if transactionID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "transaction ID is invalid (less and equal to 0)",
		})
		return
	}

	baseFlip := flip.NewBaseFlip()
	bill, err := baseFlip.ConfirmBillPayment(transactionID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get payments list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, bill)
	return
}

// --------------------------------------------------------------

func (c controller) GetExchangeRates(ctx *gin.Context) {
	var input models.GetExchangeRatesRequest
	input.CountryIsoCode = ctx.DefaultQuery("country_iso_code", "")
	input.TransactionType = strings.ToUpper(ctx.DefaultQuery("transaction_type", "C2C"))

	baseFlip := flip.NewBaseFlip()
	exchangeRates, err := baseFlip.GetExchangeRate(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get exchange rate list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, exchangeRates)
	return
}

func (c controller) GetFormData(ctx *gin.Context) {
	var input models.GetFormDataRequest
	input.CountryIsoCode = ctx.DefaultQuery("country_iso_code", "")
	input.TransactionType = strings.ToUpper(ctx.DefaultQuery("transaction_type", "C2C"))

	baseFlip := flip.NewBaseFlip()
	exchangeRates, err := baseFlip.GetFormData(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get form data",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, exchangeRates)
	return
}

func (c controller) CreateIntTransferC2X(ctx *gin.Context) {
	var input models.CreateInternationalTransferC2CC2BRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	idempotencyKey := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(idempotencyKey)
	internationalTransfer, idemKey, err := baseFlip.CreateInternationalTransferC2CC2B(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create international transfer C2C/C2B",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idemKey,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, internationalTransfer)
	return
}

func (c controller) CreateIntTransferB2X(ctx *gin.Context) {
	var input models.CreateInternationalTransferB2XRequest
	input.DestinationCountry = ctx.DefaultPostForm("destination_country", "")
	input.SourceCountry = ctx.DefaultPostForm("source_country", "")
	input.TransactionType = ctx.DefaultPostForm("transaction_type", "")
	input.Amount = ctx.DefaultPostForm("amount", "")
	input.AttachmentType = ctx.DefaultPostForm("attachment_type", "")
	input.BeneficiaryAccountNumber = ctx.DefaultPostForm("beneficiary_account_number", "")
	input.BeneficiaryAchCode = ctx.DefaultPostForm("beneficiary_ach_code", "")
	input.BeneficiaryAddress = ctx.DefaultPostForm("beneficiary_address", "")
	input.BeneficiaryBankId = ctx.DefaultPostForm("beneficiary_bank_id", "")
	input.BeneficiaryBankName = ctx.DefaultPostForm("beneficiary_bank_name", "")
	input.BeneficiaryBranchNumber = ctx.DefaultPostForm("beneficiary_branch_number", "")
	input.BeneficiaryBsbNumber = ctx.DefaultPostForm("beneficiary_bsb_number", "")
	input.BeneficiaryCity = ctx.DefaultPostForm("beneficiary_city", "")
	input.BeneficiaryDocumentReferenceNumber = ctx.DefaultPostForm("beneficiary_document_reference_number", "")
	input.BeneficiaryEmail = ctx.DefaultPostForm("beneficiary_email", "")
	input.BeneficiaryFullName = ctx.DefaultPostForm("beneficiary_full_name", "")
	input.BeneficiaryIban = ctx.DefaultPostForm("beneficiary_iban", "")
	input.BeneficiaryIdExpirationDate = ctx.DefaultPostForm("beneficiary_id_expiration_date", "")
	input.BeneficiaryIfsCode = ctx.DefaultPostForm("beneficiary_ifs_code", "")
	input.BeneficiaryIdNumber = ctx.DefaultPostForm("beneficiary_id_number", "")
	input.BeneficiaryMsisdn = ctx.DefaultPostForm("beneficiary_msisdn", "")
	input.BeneficiaryNationality = ctx.DefaultPostForm("beneficiary_nationality", "")
	input.BeneficiaryPostalCode = ctx.DefaultPostForm("beneficiary_postal_code", "")
	input.BeneficiaryProvince = ctx.DefaultPostForm("beneficiary_province", "")
	input.BeneficiaryRelationship = ctx.DefaultPostForm("beneficiary_relationship", "")
	input.BeneficiaryRemittancePurposes = ctx.DefaultPostForm("beneficiary_remittance_purposes", "")
	input.BeneficiarySortCode = ctx.DefaultPostForm("beneficiary_sort_code", "")
	input.BeneficiarySourceOfFunds = ctx.DefaultPostForm("beneficiary_source_of_funds", "")

	updateFileData, err := getFiles(ctx, "attachment_data", "attachment_data")
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message:      "image input is invalid",
			ErrorMessage: err.Error(),
		})
		return
	}

	inputFile := models.CreateInternationalTransferB2XFileRequest{
		AttachmentData: *updateFileData,
	}

	idempotencyKey := uuid.New().String()

	baseFlip := flip.NewBaseFlip()
	baseFlip.SetIdempotencyKey(idempotencyKey)
	internationalTransfer, idemKey, err := baseFlip.CreateInternationalTransferB2CB2B(input, inputFile)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:        "failed to create international transfer B2C/B2B",
			ErrorMessage:   err.Error(),
			ErrorDetails:   baseFlip.GetErrorDetails(),
			IdempotencyKey: idemKey,
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, internationalTransfer)
	return
}

func (c controller) GetIntTransferDetail(ctx *gin.Context) {
	transactionIDParam := ctx.Param("transactionID")
	transactionID := cast.ToInt64(transactionIDParam)
	if transactionID <= 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, ErrorResponse{
			Message: "transaction ID is invalid (less and equal to 0)",
		})
		return
	}

	baseFlip := flip.NewBaseFlip()
	internationalTransfer, err := baseFlip.GetInternationalTransfer(transactionID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get international transfer detail",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, internationalTransfer)
	return
}

func (c controller) GetIntTransferList(ctx *gin.Context) {
	var input models.GetAllInternationalTransferRequest
	input.Pagination = ctx.DefaultQuery("pagination", "10")
	input.Page = ctx.DefaultQuery("page", "1")
	input.StartDate = ctx.DefaultQuery("start_date", "")
	input.EndDate = ctx.DefaultQuery("end_date", "")
	input.SortBy = ctx.DefaultQuery("sort_by", "id")
	input.SortType = ctx.DefaultQuery("sort_type", "sort_desc")

	baseFlip := flip.NewBaseFlip()
	internationalTransfer, err := baseFlip.GetAllInternationalTransfer(input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get international transfer list",
			ErrorMessage: err.Error(),
			ErrorDetails: baseFlip.GetErrorDetails(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, internationalTransfer)
	return
}
