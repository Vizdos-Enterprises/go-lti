package lti_testadapters

import "net/http"

type FakeRedirect struct {
	swapToken   string
	didRedirect bool
}

func (c *FakeRedirect) RedirectAfterLaunch(w http.ResponseWriter, r *http.Request, swapToken string) {
	c.swapToken = swapToken
	c.didRedirect = true
}

func (c *FakeRedirect) DidRedirect() bool {
	return c.didRedirect
}

func (c *FakeRedirect) HasToken(expectToken string) bool {
	return c.swapToken == expectToken
}

func (c *FakeRedirect) HasSwapToken() bool {
	return c.swapToken != ""
}
