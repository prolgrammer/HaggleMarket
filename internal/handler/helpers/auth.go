package helpers

import "github.com/gin-gonic/gin"

const (
	UserIDKey      = "userID"
	UserIsStoreKey = "userIsStore"
	UserIsAdminKey = "userIsAdmin"
	UserName       = "userName"
	UserEmail      = "userEmail"
)

func UserIDContext(c *gin.Context) (uint, bool) {
	id, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	idValue, ok := id.(uint)
	if !ok {
		return 0, false
	}
	return idValue, true
}

func UserIsAdminContext(c *gin.Context) (bool, bool) {
	isAdmin, exists := c.Get(UserIsAdminKey)
	if !exists {
		return false, false
	}
	isAdminValue, ok := isAdmin.(bool)
	if !ok {
		return false, false
	}
	return isAdminValue, true
}

func UserIsStoreContext(c *gin.Context) (bool, bool) {
	isStore, exists := c.Get(UserIsStoreKey)
	if !exists {
		return false, false
	}
	isStoreValue, ok := isStore.(bool)
	if !ok {
		return false, false
	}
	return isStoreValue, true
}

func UserNameContext(c *gin.Context) (string, bool) {
	name, exists := c.Get(UserName)
	if !exists {
		return "", false
	}
	nameValue, ok := name.(string)
	if !ok {
		return "", false
	}
	return nameValue, true
}

func UserEmailContext(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmail)
	if !exists {
		return "", false
	}
	emailValue, ok := email.(string)
	if !ok {
		return "", false
	}
	return emailValue, true
}
