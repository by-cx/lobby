package main

import (
	"strings"

	"github.com/labstack/echo"
)

func TokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip selected paths

		tokenHeader := c.Request().Header.Get("Authorization")
		token := strings.Replace(tokenHeader, "Token ", "", -1)

		if token != config.Token || config.Token == "" {
			return c.JSONPretty(403, map[string]string{"message": "access denied"}, " ")
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
