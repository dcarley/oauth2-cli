package main_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2"
)

func EphemeralPort() (int, error) {
	const proto = "tcp"

	addr, err := net.ResolveTCPAddr(proto, ":0")
	if err != nil {
		return 0, err
	}
	sock, err := net.ListenTCP(proto, addr)
	if err != nil {
		return 0, err
	}
	defer sock.Close()

	return sock.Addr().(*net.TCPAddr).Port, nil
}

var _ = Describe("Main", func() {
	var (
		args    []string
		session *gexec.Session
		server  *ghttp.Server
		authURL *url.URL
	)

	BeforeEach(func() {
		args = []string{"-scope", "public"}
		server = ghttp.NewServer()
	})

	JustBeforeEach(func() {
		port, err := EphemeralPort()
		Expect(err).ToNot(HaveOccurred())

		authURLBase := server.URL() + "/oauth/authorize"
		args = append(args, []string{
			"-port", fmt.Sprintf("%d", port),
			"-auth", authURLBase,
			"-token", server.URL() + "/oauth/token",
			"-id", "123",
			"-secret", "abc",
		}...)
		command := exec.Command(cmdPath, args...)

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		var authURLRaw string
		re := regexp.MustCompile(regexp.QuoteMeta(authURLBase) + `.+`)
		Eventually(func() bool {
			authURLRaw = string(re.Find(session.Out.Contents()))
			return authURLRaw != ""
		}).Should(BeTrue(), "couldn't find auth URL in STDOUT")

		authURL, err = url.Parse(authURLRaw)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("successful token exchange", func() {
		const (
			expectedToken = "mytoken"
			expectedCode  = "mycode"
		)

		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.VerifyFormKV("code", expectedCode),
				ghttp.RespondWithJSONEncoded(http.StatusOK, oauth2.Token{
					AccessToken: expectedToken,
					TokenType:   "Bearer",
				}),
			))
		})

		It("should output access token", func() {
			callbackURL, err := url.Parse(authURL.Query().Get("redirect_uri"))
			Expect(err).ToNot(HaveOccurred())

			params := callbackURL.Query()
			params.Set("code", expectedCode)
			callbackURL.RawQuery = params.Encode()

			resp, err := http.Get(callbackURL.String())
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK), "got body: %s", body)
			Expect(string(body)).To(Equal(fmt.Sprintf(`{
  "access_token": "%s",
  "token_type": "Bearer",
  "expiry": "0001-01-01T00:00:00Z"
}`, expectedToken)))

			gexec.Terminate()
			Eventually(session).Should(gexec.Exit(143))
		})
	})

	Describe("unsuccessful token exchange", func() {
		const (
			expectedResponse = "bad things happened"
		)

		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.RespondWith(http.StatusBadRequest, expectedResponse),
			))
		})

		It("should output error", func() {
			callbackURL, err := url.Parse(authURL.Query().Get("redirect_uri"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := http.Get(callbackURL.String())
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable), "got body: %s", body)
			Expect(string(body)).To(Equal(`Exchange error: oauth2: cannot fetch token: 400 Bad Request
Response: bad things happened
`))

			gexec.Terminate()
			Eventually(session).Should(gexec.Exit(143))
		})
	})
})
