package guestoffer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGuestOfferRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Guest Offer Repository Suite")
}
