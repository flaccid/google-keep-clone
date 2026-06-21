package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("media", func() {
	Description("The media service handles attachment downloads.")

	Method("download", func() {
		Description("Downloads an attachment.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("attachmentId", String, "The ID of the attachment.")
			Required("noteId", "attachmentId")
		})

		Result(Bytes, "The attachment content.")

		HTTP(func() {
			GET("/v1/notes/{noteId}/attachments/{attachmentId}")
			Response(StatusOK, func() {
				ContentType("application/octet-stream")
			})
		})
	})
})
