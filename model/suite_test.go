package model

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScheme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "model suite")
}
