package user

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Service Suite")
}
