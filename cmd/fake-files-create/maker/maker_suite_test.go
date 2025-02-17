package maker_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maker Suite")
}
