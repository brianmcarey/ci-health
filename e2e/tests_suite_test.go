package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"

	"github.com/fgimenez/cihealth/pkg/runner"
	"github.com/fgimenez/cihealth/pkg/stats"
	"github.com/fgimenez/cihealth/pkg/types"
)

func TestCIHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cihealth Suite")
}

var (
	tokenPath string
)

const (
	source   = "kubevirt/kubevirt"
	dataDays = 7
)

var _ = BeforeSuite(func() {
	tokenPath = os.Getenv("GITHUB_TOKEN_PATH")
	if tokenPath == "" {
		Fail("Please specify an OAuth2 token in the env var GITHUB_TOKEN_PATH")
	}
})

var _ = Describe("cihealth stats", func() {
	It("Retrieves data from github", func() {
		opt := &types.Options{
			TokenPath: tokenPath,
			Source:    source,
			DataDays:  dataDays,
			LogLevel:  "debug",
		}

		path, err := runner.Run(opt)
		Expect(err).ToNot(HaveOccurred())

		contents, err := ioutil.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())

		results := stats.Results{}
		err = yaml.Unmarshal(contents, &results)

		Expect(results.DataDays).To(Equal(dataDays))
		Expect(results.Source).To(Equal(source))
	})
})
