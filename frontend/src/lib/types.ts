export interface Note {
  name?: string
  createTime?: string
  updateTime?: string
  trashTime?: string | null
  trashed?: boolean
  title?: string
  body?: Section
  pinned?: boolean
  archived?: boolean
  color?: string
  labels?: string[]
}

export interface NoteRequest {
  title?: string
  body?: Section
  pinned?: boolean
  archived?: boolean
  color?: string
  labels?: string[]
}

export interface Section {
  text?: TextContent
  list?: ListContent
}

export interface TextContent {
  text?: string
}

export interface ListContent {
  listItems?: ListItem[]
}

export interface ListItem {
  text?: TextContent
  checked?: boolean
  childListItems?: ListItem[]
}

export interface ListNotesResponse {
  notes: Note[]
  nextPageToken?: string | null
}

export interface Label {
  name?: string
  displayName?: string
}
