package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("notes", func() {
	Description("The notes service handles CRUD operations for notes.")

	Method("create", func() {
		Description("Creates a new note.")

		Payload(func() {
			Attribute("note", NoteRequest, "The note to create.")
		})

		Result(Note, "The newly created note.")

		HTTP(func() {
			POST("/v1/notes")
			Body("note")
			Response(StatusCreated)
		})
	})

	Method("get", func() {
		Description("Gets a note by id.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The requested note.")

		HTTP(func() {
			GET("/v1/notes/{id}")
		})
	})

	Method("list", func() {
		Description("Lists notes with optional filtering and pagination.")

		Payload(func() {
			Attribute("pageSize", Int, "The maximum number of results to return.")
			Attribute("pageToken", String, "The previous page's nextPageToken field.")
			Attribute("filter", String, "Filter for list results. Valid fields: createTime, updateTime, trashTime, trashed.")
		})

		Result(ListNotesResponse, "A page of notes.")

		HTTP(func() {
			GET("/v1/notes")
			Param("pageSize")
			Param("pageToken")
			Param("filter")
		})
	})

	Method("update", func() {
		Description("Updates a note. Fields not specified are left unchanged.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Attribute("note", NoteRequest, "The note fields to update.")
			Required("id", "note")
		})

		Result(Note, "The updated note.")

		HTTP(func() {
			PATCH("/v1/notes/{id}")
			Body("note")
		})
	})

	Method("delete", func() {
		Description("Deletes a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note to delete.")
			Required("id")
		})

		HTTP(func() {
			DELETE("/v1/notes/{id}")
			Response(StatusNoContent)
		})
	})

	Method("pin", func() {
		Description("Pins a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The pinned note.")

		HTTP(func() {
			POST("/v1/notes/{id}:pin")
		})
	})

	Method("unpin", func() {
		Description("Unpins a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The unpinned note.")

		HTTP(func() {
			POST("/v1/notes/{id}:unpin")
		})
	})

	Method("archive", func() {
		Description("Archives a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The archived note.")

		HTTP(func() {
			POST("/v1/notes/{id}:archive")
		})
	})

	Method("unarchive", func() {
		Description("Unarchives a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The unarchived note.")

		HTTP(func() {
			POST("/v1/notes/{id}:unarchive")
		})
	})

	Method("trash", func() {
		Description("Trashes a note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The trashed note.")

		HTTP(func() {
			POST("/v1/notes/{id}:trash")
		})
	})

	Method("restore", func() {
		Description("Restores a trashed note.")

		Payload(func() {
			Attribute("id", String, "The ID of the note.")
			Required("id")
		})

		Result(Note, "The restored note.")

		HTTP(func() {
			POST("/v1/notes/{id}:restore")
		})
	})
})
