package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nazarslota/unotes/auth/internal/service/oauth2"
)

type oAuth2SignUpUserModel struct {
	Username string `json:"username" validate:"required,min=4,max=32" example:"username"`
	Password string `json:"password" validate:"required,min=8,max=64" example:"password"`
}

// @Summary     oAuth2 sign up
// @Description create account
// @Tags        oAuth2
// @Accept      json
// @Produce     json
// @Param       input body oAuth2SignUpUserModel true "account info"
// @Success     204
// @Failure     400     {object} errors.HTTPError
// @Failure     409     {object} errors.HTTPError
// @Failure     500     {object} errors.HTTPError
// @Failure     default {object} errors.HTTPError
// @Router      /oauth2/sign-up [post]
func (h *Handler) oAuth2SignUp(c echo.Context) error {
	input := new(oAuth2SignUpUserModel)
	if err := c.Bind(input); err != nil {
		return err
	}

	if err := c.Validate(input); err != nil {
		return err
	}

	request := &oauth2.SignUpRequest{
		Username: input.Username,
		Password: input.Password,
	}

	_, err := h.services.OAuth2Service.SignUpRequestHandler.Handler(c.Request().Context(), request)
	if errors.Is(err, oauth2.ErrSignUpUserAlreadyExist) {
		return echo.NewHTTPError(http.StatusConflict, "A user with this username already exists").SetInternal(err)
	} else if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type oAuth2SignInUserModel struct {
	Username string `json:"username" validate:"required,min=4,max=32" example:"username"`
	Password string `json:"password" validate:"required,min=8,max=64" example:"password"`
}

type oAuth2SignInUserResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary     oAuth2 sign in
// @Description sign in
// @Tags        oAuth2
// @Accept      json
// @Produce     json
// @Param       input   body     oAuth2SignInUserModel true "account info"
// @Success     200     {object} oAuth2SignInUserResult
// @Failure     400     {object} errors.HTTPError
// @Failure     404     {object} errors.HTTPError
// @Failure     500     {object} errors.HTTPError
// @Failure     default {object} errors.HTTPError
// @Router      /oauth2/sign-in [post]
func (h *Handler) oAuth2SignIn(c echo.Context) error {
	input := new(oAuth2SignInUserModel)
	if err := c.Bind(input); err != nil {
		return err
	}

	if err := c.Validate(input); err != nil {
		return err
	}

	request := &oauth2.SignInRequest{
		Username: input.Username,
		Password: input.Password,
	}

	result, err := h.services.OAuth2Service.SingInRequestHandler.Handle(c.Request().Context(), request)
	if errors.Is(err, oauth2.ErrSignInUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User with that username was not found")
	} else if errors.Is(err, oauth2.ErrSignInInvalidPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid password")
	} else if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	return c.JSON(http.StatusOK, oAuth2SignInUserResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

type oAuth2LogOutModel struct {
	AccessToken string `json:"access_token" validate:"required,jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// @Summary     oAuth2 sign out
// @Description sign out
// @Tags        oAuth2
// @Accept      json
// @Produce     json
// @Param       input body oAuth2LogOutModel true "account info"
// @Success     204
// @Failure     400     {object} errors.HTTPError
// @Failure     500     {object} errors.HTTPError
// @Failure     default {object} errors.HTTPError
// @Router      /oauth2/sign-out [post]
func (h *Handler) oAuth2SignOut(c echo.Context) error {
	input := new(oAuth2LogOutModel)
	if err := c.Bind(input); err != nil {
		return err
	}

	if err := c.Validate(input); err != nil {
		return err
	}

	request := &oauth2.SignOutRequest{
		AccessToken: input.AccessToken,
	}

	_, err := h.services.OAuth2Service.SignOutRequestHandler.Handle(c.Request().Context(), request)
	if errors.Is(err, oauth2.ErrSignOutInvalidOrExpiredToken) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired token")
	} else if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type oAuth2RefreshModel struct {
	RefreshToken string `json:"refresh_token" validator:"required,jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type oAuth2RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary     oAuth2 refresh
// @Description refresh
// @Tags        oAuth2
// @Accept      json
// @Produce     json
// @Param       input   body     oAuth2RefreshModel true "account info"
// @Success     200     {object} oAuth2RefreshResult
// @Failure     400     {object} errors.HTTPError
// @Failure     500     {object} errors.HTTPError
// @Failure     default {object} errors.HTTPError
// @Router      /oauth2/refresh [post]
func (h *Handler) oAuth2Refresh(c echo.Context) error {
	input := new(oAuth2RefreshModel)
	if err := c.Bind(input); err != nil {
		return err
	}

	if err := c.Validate(input); err != nil {
		return err
	}

	request := &oauth2.RefreshRequest{
		RefreshToken: input.RefreshToken,
	}

	result, err := h.services.OAuth2Service.RefreshRequestHandler.Handle(c.Request().Context(), request)
	if errors.Is(err, oauth2.ErrRefreshInvalidOrExpiredToken) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired token")
	} else if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	return c.JSON(http.StatusOK, oAuth2RefreshResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}
