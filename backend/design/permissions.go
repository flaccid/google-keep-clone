package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("permissions", func() {
	Description("The permissions service handles batch operations on note permissions.")

	Method("batchCreate", func() {
		Description("Creates one or more permissions on a note.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("batchCreatePermissionsRequest", BatchCreatePermissionsRequest, "The permissions to create.")
			Required("noteId", "batchCreatePermissionsRequest")
		})

		Result(ArrayOf(Permission), "The created permissions.")

		HTTP(func() {
			POST("/v1/notes/{noteId}/permissions:batchCreate")
			Body("batchCreatePermissionsRequest")
		})
	})

	Method("batchDelete", func() {
		Description("Deletes one or more permissions on a note.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("batchDeletePermissionsRequest", BatchDeletePermissionsRequest, "The permissions to delete.")
			Required("noteId", "batchDeletePermissionsRequest")
		})

		HTTP(func() {
			POST("/v1/notes/{noteId}/permissions:batchDelete")
			Body("batchDeletePermissionsRequest")
			Response(StatusNoContent)
		})
	})
})
