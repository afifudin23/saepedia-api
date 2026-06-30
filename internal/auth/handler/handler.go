package handler

import (
	"errors"
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/auth/domain"
	"github.com/afifudin23/saepedia-api/internal/auth/dto"
	userdomain "github.com/afifudin23/saepedia-api/internal/user/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc domain.AuthUsecase
}

func New(uc domain.AuthUsecase) *Handler {
	return &Handler{uc: uc}
}

// Register godoc
// @Summary  Registrasi user baru
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      dto.RegisterRequest  true  "Data registrasi (login pakai email)"
// @Success  201   {object}  response.Response
// @Failure  400   {object}  response.Response
// @Failure  409   {object}  response.Response
// @Router   /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.uc.Register(c.Request.Context(), domain.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Roles:    req.Roles,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrEmailExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}

	response.Success(c, http.StatusCreated, "user registered", dto.ToAuthResponse(result))
}

// Login godoc
// @Summary  Login (pakai email)
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      dto.LoginRequest  true  "Kredensial login"
// @Success  200   {object}  response.Response
// @Failure  401   {object}  response.Response
// @Router   /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.uc.Login(c.Request.Context(), domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}

	msg := "login success"
	if result.NeedRoleSelection {
		msg = "login success, please select an active role"
	}
	response.Success(c, http.StatusOK, msg, dto.ToAuthResponse(result))
}

// Logout godoc
// @Summary   Logout (invalidasi token via denylist)
// @Tags      Auth
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Failure   401  {object}  response.Response
// @Router    /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if err := h.uc.Logout(c.Request.Context(), middleware.JTI(c), middleware.TokenExp(c)); err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "logout success, token has been invalidated", nil)
}

// SelectRole godoc
// @Summary   Pilih role aktif (terbitkan token baru)
// @Tags      Auth
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.SelectRoleRequest  true  "Role yang dipilih"
// @Success   200   {object}  response.Response
// @Failure   403   {object}  response.Response
// @Router    /auth/select-role [post]
func (h *Handler) SelectRole(c *gin.Context) {
	var req dto.SelectRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.uc.SelectRole(c.Request.Context(), middleware.UserID(c), req.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoleNotOwned):
			response.Forbidden(c, err.Error())
		case errors.Is(err, userdomain.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.Success(c, http.StatusOK, "active role selected", dto.ToAuthResponse(result))
}

// Me godoc
// @Summary   Profil user (roles + active role)
// @Tags      Auth
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	user, err := h.uc.Me(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}

	response.Success(c, http.StatusOK, "", dto.ToProfileResponse(user, middleware.ActiveRole(c)))
}

// BalanceSummary godoc
// @Summary   Ringkasan saldo lintas role
// @Tags      Auth
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /auth/balance-summary [get]
func (h *Handler) BalanceSummary(c *gin.Context) {
	summary, err := h.uc.BalanceSummary(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", summary)
}
