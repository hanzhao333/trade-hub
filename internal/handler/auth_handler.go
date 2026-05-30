package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/internal/middleware"
	"github.com/zhanghanzhao/trade-hub/internal/service"
	"github.com/zhanghanzhao/trade-hub/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.auth.Register(req.Email, req.Password); err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			response.Fail(c, 409, 409, "email already exists")
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"registered": true})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	token, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetUint(middleware.UserIDKey)
	u, err := h.auth.GetProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, u)
}
