package design

import (
	. "goa.design/goa/v3/dsl"
)

var ColorValue = Type("ColorValue", String, func() {
	Description("The color or theme of a note. Matches Google Keep's color palette and background themes.")
	Enum("DEFAULT", "RED", "ORANGE", "YELLOW", "GREEN", "TEAL", "BLUE", "CERULEAN", "PURPLE", "PINK", "BROWN", "GRAY",
		"THEME_SHORE", "THEME_BLOOM", "THEME_PLUM", "THEME_NIGHT", "THEME_BAMBOO", "THEME_CANDY", "THEME_SUNSET", "THEME_OCEAN")
})

var Role = Type("Role", String, func() {
	Description("The role granted by a permission.")
	Enum("OWNER", "WRITER")
})

var TextContent = Type("TextContent", func() {
	Description("The block of text for a single text section or list item.")
	Attribute("text", String, "The text content.")
})

var ListItem = Type("ListItem", func() {
	Description("A single list item in a note's list.")
	Attribute("text", TextContent, "The text of this item.")
	Attribute("checked", Boolean, "Whether this item has been checked off.")
	Attribute("childListItems", ArrayOf("ListItem"), "Nested list items. Only one level of nesting is allowed.")
})

var ListContent = Type("ListContent", func() {
	Description("The list of items for a single list note.")
	Attribute("listItems", ArrayOf(ListItem), "The items in the list.")
})

var Section = Type("Section", func() {
	Description("The content of the note. Either text or list content.")
	Attribute("text", TextContent, "Used if this section's content is a block of text.")
	Attribute("list", ListContent, "Used if this section's content is a list.")
})

var Attachment = Type("Attachment", func() {
	Description("An attachment to a note.")
	Attribute("name", String, "The resource name of the attachment.")
	Attribute("mimeType", ArrayOf(String), "The MIME types in which the attachment is available.")
})

var User = Type("User", func() {
	Description("Describes a single user.")
	Attribute("email", String, "The user's email.")
})

var Group = Type("Group", func() {
	Description("Describes a single group.")
	Attribute("email", String, "The group email.")
})

var Family = Type("Family", func() {
	Description("Describes a single Google Family. Empty type.")
})

var Permission = Type("Permission", func() {
	Description("A single permission on the note. Associates a member with a role.")
	Attribute("name", String, "Output only. The resource name.")
	Attribute("role", Role, "The role granted by this permission.")
	Attribute("email", String, "The email associated with the member.")
	Attribute("deleted", Boolean, "Output only. Whether this member has been deleted.")
	Attribute("user", User, "Output only. The user to whom this role applies.")
	Attribute("group", Group, "Output only. The group to which this role applies.")
	Attribute("family", Family, "Output only. The Google Family to which this role applies.")
})

var Note = Type("Note", func() {
	Description("A single note.")
	Attribute("name", String, "Output only. The resource name of this note (e.g. 'notes/{uuid}').")
	Attribute("createTime", String, "Output only. When this note was created (RFC3339 format).")
	Attribute("updateTime", String, "Output only. When this note was last modified (RFC3339 format).")
	Attribute("trashTime", String, "Output only. When this note was trashed. Only set if trashed.")
	Attribute("trashed", Boolean, "Output only. True if this note has been trashed.")
	Attribute("title", String, "The title of the note. Length must be less than 1,000 characters.")
	Attribute("body", Section, "The body of the note.")
	Attribute("attachments", ArrayOf(Attachment), "Output only. The attachments attached to this note.")
	Attribute("permissions", ArrayOf(Permission), "Output only. The list of permissions set on the note.")
	Attribute("pinned", Boolean, "Whether the note is pinned.")
	Attribute("archived", Boolean, "Whether the note is archived.")
	Attribute("color", ColorValue, "The color of the note.")
	Attribute("labels", ArrayOf(String), "The labels assigned to this note.")
})

var NoteRequest = Type("NoteRequest", func() {
	Description("Request payload for creating or updating a note.")
	Attribute("title", String, "The title of the note.", func() {
		MaxLength(1000)
	})
	Attribute("body", Section, "The body of the note.")
	Attribute("pinned", Boolean, "Whether the note is pinned.")
	Attribute("archived", Boolean, "Whether the note is archived.")
	Attribute("color", ColorValue, "The color of the note.")
	Attribute("labels", ArrayOf(String), "The labels assigned to this note.")
})

var Label = Type("Label", func() {
	Description("A label that can be assigned to notes.")
	Attribute("name", String, "The resource name of the label (e.g. 'labels/{uuid}').")
	Attribute("displayName", String, "The display name of the label.")
})

var ListNotesResponse = Type("ListNotesResponse", func() {
	Description("The response when listing a page of notes.")
	Attribute("notes", ArrayOf(Note), "A page of notes.")
	Attribute("nextPageToken", String, "Next page's pageToken field.")
})

var BatchCreatePermissionsRequest = Type("BatchCreatePermissionsRequest", func() {
	Description("Request to create one or more permissions on a note.")
	Attribute("requests", ArrayOf(CreatePermissionRequest), "The permission requests to create.")
})

var CreatePermissionRequest = Type("CreatePermissionRequest", func() {
	Description("A request to create a single permission.")
	Attribute("role", Role, "The role to grant.")
	Attribute("email", String, "The email of the user to grant the permission to.")
})

var BatchDeletePermissionsRequest = Type("BatchDeletePermissionsRequest", func() {
	Description("Request to delete one or more permissions on a note.")
	Attribute("names", ArrayOf(String), "The resource names of the permissions to delete.")
})
