package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var cmdPath string

var _ = BeforeSuite(func() {
	var err error
	cmdPath, err = gexec.Build("github.com/dcarley/oauth2-cli")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestOauthCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Oauth2Cli Suite")
}
