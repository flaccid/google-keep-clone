package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("media", func() {
	Description("The media service handles attachment uploads and downloads.")

	Method("upload", func() {
		Description("Uploads an attachment to a note.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("contentType", String, "The MIME type of the attachment.", func() {
				Example("image/png")
			})
			Attribute("data", Bytes, "The attachment data.")
			Required("noteId", "contentType", "data")
		})

		Result(Attachment, "The created attachment.")

		HTTP(func() {
			POST("/v1/notes/{noteId}/attachments")
			Header("contentType:Content-Type")
			Body("data")
			Response(StatusCreated)
		})
	})

	Method("download", func() {
		Description("Downloads an attachment.")

		Payload(func() {
			Attribute("noteId", String, "The ID of the note.")
			Attribute("attachmentId", String, "The ID of the attachment.")
			Attribute("mimeType", String, "The requested MIME type. Must be one of the attachment's mimeType values.")
			Attribute("alt", String, "The alt query parameter. Use 'media' to return raw bytes; omit to return Attachment metadata.")
			Required("noteId", "attachmentId")
		})

		Result(Bytes, "The attachment content.")

		HTTP(func() {
			GET("/v1/notes/{noteId}/attachments/{attachmentId}")
			Param("mimeType")
			Param("alt")
			Response(StatusOK, func() {
				ContentType("application/octet-stream")
			})
		})
	})
})
