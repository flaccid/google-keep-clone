package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("keep", func() {
	Title("Google Keep Clone API")
	Description("A clone of the Google Keep REST API, designed to mirror the official API surface.")
	Version("1.0")

	Server("keep-server", func() {
		Description("The Google Keep Clone server.")
		Host("localhost", func() {
			Description("Local development host.")
			URI("http://localhost:8080")
		})
	})
})
