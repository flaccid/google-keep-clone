package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("labels", func() {
	Description("The labels service manages label resources.")

	Method("list", func() {
		Description("Lists all labels.")

		Result(ArrayOf(Label), "The list of labels.")

		HTTP(func() {
			GET("/v1/labels")
		})
	})

	Method("create", func() {
		Description("Creates a new label.")

		Payload(func() {
			Attribute("displayName", String, "The display name of the label.")
			Required("displayName")
		})

		Result(Label, "The created label.")

		HTTP(func() {
			POST("/v1/labels")
		})
	})

	Method("update", func() {
		Description("Updates the display name of a label.")

		Payload(func() {
			Attribute("id", String, "The ID of the label to update.")
			Attribute("displayName", String, "The new display name for the label.")
			Required("id", "displayName")
		})

		Result(Label, "The updated label.")

		HTTP(func() {
			PATCH("/v1/labels/{id}")
		})
	})

	Method("delete", func() {
		Description("Deletes a label.")

		Payload(func() {
			Attribute("id", String, "The ID of the label to delete.")
			Required("id")
		})

		HTTP(func() {
			DELETE("/v1/labels/{id}")
			Response(StatusNoContent)
		})
	})
})
