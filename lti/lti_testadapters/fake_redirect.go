package lti_testadapters

import "net/http"

type FakeRedirect struct {
	passedToken string
	didRedirect bool
}

func (c *FakeRedirect) RedirectAfterLaunch(w http.ResponseWriter, r *http.Request, token string) {
	c.passedToken = token
	c.didRedirect = true
}

func (c *FakeRedirect) DidRedirect() bool {
	return c.didRedirect
}

func (c *FakeRedirect) HasToken(expectToken string) bool {
	return c.passedToken == expectToken
}
