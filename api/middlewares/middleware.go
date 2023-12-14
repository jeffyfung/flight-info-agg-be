package middlewares

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jeffyfung/flight-info-agg/config"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
	jwt.RegisteredClaims
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		authHeader, ok := c.Request().Header["Authorization"]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing Authorization header")
		}

		claims := &Claims{}
		jwtTokenStr := authHeader[0]
		_, err := jwt.ParseWithClaims(jwtTokenStr, claims, func(token *jwt.Token) (any, error) {
			return config.Cfg.Server.JwtSecret, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		// if !jwtToken.Valid {
		// 	return echo.NewHTTPError(http.StatusUnauthorized)
		// }

		c.Set("user", model.UserPublicInfo{Email: claims.Email, Role: claims.Role})

		return next(c)
	}
}
