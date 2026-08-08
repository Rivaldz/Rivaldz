package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// authUserID returns the authenticated user ID from the request context.
func authUserID(ctx *fiber.Ctx) (string, bool) {
	userID, ok := ctx.Locals("userID").(string)

	return userID, ok
}

// bindBody parses and validates the request body into dst.
func (r *V1) bindBody(ctx *fiber.Ctx, op string, dst any) bool {
	if err := ctx.BodyParser(dst); err != nil {
		r.l.Error(err, "restapi - v1 - "+op)

		return false
	}

	if err := r.v.Struct(dst); err != nil {
		r.l.Error(err, "restapi - v1 - "+op)

		return false
	}

	return true
}

// unauthorizedResponse writes a 401 error response.
func unauthorizedResponse(ctx *fiber.Ctx) error {
	return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
}

// badRequestResponse writes a 400 error response.
func badRequestResponse(ctx *fiber.Ctx) error {
	return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
}
