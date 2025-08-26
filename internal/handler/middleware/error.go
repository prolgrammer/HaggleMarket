package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/gin-gonic/gin"
)

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr apperror.AppError
			if errors.As(err, &appErr) {
				statusCode := errorStatus(appErr.Code())
				resp := gin.H{
					"code":    appErr.Code(),
					"message": appErr.Message(),
				}
				if appErr.Code() == apperror.BadRequest {
					resp["details"] = appErr.Error()
				}
				c.AbortWithStatusJSON(statusCode, resp)
				return
			}

			log.Printf("Unhandled error: %v", err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    apperror.InternalError,
				"message": "An unexpected internal error occurred",
			})
			return
		}
	}
}

func errorStatus(code string) int {
	switch code {
	case apperror.NotFound:
		return http.StatusNotFound
	case apperror.DatabaseError:
		return http.StatusInternalServerError
	case apperror.DuplicateError:
		return http.StatusConflict
	case apperror.BadRequest:
		return http.StatusBadRequest
	case apperror.Unauthorized, apperror.InvalidToken:
		return http.StatusUnauthorized
	case apperror.Conflict:
		return http.StatusConflict
	case apperror.Forbidden:
		return http.StatusForbidden
	case apperror.InternalError:
		fallthrough

	default:
		return http.StatusInternalServerError
	}
}
