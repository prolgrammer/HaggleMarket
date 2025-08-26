package guestoffer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGuestOfferService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Guest Offer Service Suite")
}
