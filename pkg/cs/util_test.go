package cs_test

import (
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateUrl", func() {

	It("succeeds on a github repo URL", func() {
		urlIn := "https://github.com/codesphere-cloud/codesphere-monorepofails"
		url, err := cs.ValidateUrl(urlIn)
		Expect(url).To(Equal(urlIn))
		Expect(err).NotTo(HaveOccurred())

	})

	It("fails on unsupported scheme", func() {
		urlIn := "sftp://my-server/my-stuff"
		url, err := cs.ValidateUrl(urlIn)
		Expect(url).To(Equal(""))
		Expect(err).To(MatchError("unsupported URL scheme: sftp. Only http and https are supported"))
	})

	It("fails on invalid URL", func() {
		urlIn := "\nhttps:\n"
		url, err := cs.ValidateUrl(urlIn)
		Expect(url).To(Equal(""))
		Expect(err).To(MatchError("failed to parse URL \nhttps:\n: parse \"\\nhttps:\\n\": net/url: invalid control character in URL"))
	})

})
