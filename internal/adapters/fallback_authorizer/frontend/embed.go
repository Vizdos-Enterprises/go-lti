package pages

import (
	_ "embed"
)

//go:embed interstitial.html
var InterstitialHTML []byte

//go:embed exchange.html
var ExchangeHTML []byte

//go:embed error.html
var ErrorHTML []byte

//go:embed style.css
var Styles []byte
