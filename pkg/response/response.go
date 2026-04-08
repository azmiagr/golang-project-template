package response

import (
	"golang-project-template/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  Status      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Status struct {
	Code      int  `json:"code"`
	IsSuccess bool `json:"isSuccess"`
}

func Success(ctx *gin.Context, code int, message string, data any) {
	ctx.JSON(code, Response{
		Status: Status{
			Code:      code,
			IsSuccess: true,
		},
		Message: message,
		Data:    data,
	})
}

func Error(ctx *gin.Context, code int, message string, err error) {
	var errorData interface{}
	if err != nil {
		errorData = err.Error()
	} else {
		errorData = nil
	}

	ctx.JSON(code, Response{
		Status: Status{
			Code:      code,
			IsSuccess: false,
		},
		Message: message,
		Data:    errorData,
	})
}

func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		Error(c, appErr.Code, appErr.Message, appErr.Err)
		return
	}
	Error(c, http.StatusInternalServerError, "internal server error", err)
}
