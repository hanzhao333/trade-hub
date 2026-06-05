package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Body{Code: code, Message: msg})
}

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// 那边传了c和msg过来，这边的401以及错误状态，是我们自己定义的
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}

func Internal(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 500, msg)
}
