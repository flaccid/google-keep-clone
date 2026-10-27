package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("permissions", func() {
	Description("The permissions service handles batch operations on note permissions.")

	Method("batchCreate", func() {
		Description("Atomically creates or updates permissions. Only the actual note owner may mutate permissions. Duplicate emails are invalid; returned resource names use persisted IDs.")
		Error("invalid_request", String, "The complete batch is invalid; no changes were committed.")
		Error("forbidden", String, "Only the actual note owner may change permissions.")
		Error("not_found", String, "The note does not exist.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("batchCreatePermissionsRequest", BatchCreatePermissionsRequest, "The permissions to create.")
			Required("noteId", "batchCreatePermissionsRequest")
		})

		Result(ArrayOf(Permission), "The created permissions.")

		HTTP(func() {
			POST("/v1/notes/{noteId}/permissions:batchCreate")
			Body("batchCreatePermissionsRequest")
			Response("invalid_request", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
		})
	})

	Method("batchDelete", func() {
		Description("Atomically deletes permissions for the actual note owner. Duplicate or foreign-note names are invalid. Missing IDs are idempotent after authorization; grantor identity does not prevent owner revocation.")
		Error("invalid_request", String, "The complete batch is invalid; no changes were committed.")
		Error("forbidden", String, "Only the actual note owner may change permissions.")
		Error("not_found", String, "The note does not exist.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("batchDeletePermissionsRequest", BatchDeletePermissionsRequest, "The permissions to delete.")
			Required("noteId", "batchDeletePermissionsRequest")
		})

		HTTP(func() {
			POST("/v1/notes/{noteId}/permissions:batchDelete")
			Body("batchDeletePermissionsRequest")
			Response(StatusNoContent)
			Response("invalid_request", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
		})
	})
})
