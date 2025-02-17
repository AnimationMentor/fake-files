package maker_test

import (
	"github.com/AnimationMentor/fake-files/cmd/fake-files-create/maker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MockImage", func() {
	It("makes an PNG", func() {
		img := maker.MakePNG()
		Expect(img).NotTo(BeEmpty())
	})
	It("makes an JPEG", func() {
		img := maker.MakeJPEG()
		Expect(img).NotTo(BeEmpty())
	})
})
