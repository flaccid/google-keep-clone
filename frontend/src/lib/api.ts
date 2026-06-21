import type { Note, NoteRequest, ListNotesResponse, Label } from "./types"

const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  })
  if (!res.ok) {
    const text = await res.text().catch(() => "")
    throw new Error(`${res.status} ${res.statusText}: ${text}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  notes: {
    list: (pageSize?: number, pageToken?: string, search?: string) => {
      const params = new URLSearchParams()
      if (pageSize) params.set("pageSize", String(pageSize))
      if (pageToken) params.set("pageToken", pageToken)
      if (search) params.set("search", search)
      const qs = params.toString()
      return request<ListNotesResponse>(`/v1/notes${qs ? "?" + qs : ""}`)
    },
    get: (id: string) => request<Note>(`/v1/notes/${id}`),
    create: (note: NoteRequest) =>
      request<Note>("/v1/notes", { method: "POST", body: JSON.stringify(note) }),
    update: (id: string, note: NoteRequest) =>
      request<Note>(`/v1/notes/${id}`, { method: "PATCH", body: JSON.stringify(note) }),
    delete: (id: string) =>
      request<void>(`/v1/notes/${id}`, { method: "DELETE" }),
    pin: (id: string) => request<Note>(`/v1/notes/${id}:pin`, { method: "POST" }),
    unpin: (id: string) => request<Note>(`/v1/notes/${id}:unpin`, { method: "POST" }),
    archive: (id: string) => request<Note>(`/v1/notes/${id}:archive`, { method: "POST" }),
    unarchive: (id: string) => request<Note>(`/v1/notes/${id}:unarchive`, { method: "POST" }),
    trash: (id: string) => request<Note>(`/v1/notes/${id}:trash`, { method: "POST" }),
    restore: (id: string) => request<Note>(`/v1/notes/${id}:restore`, { method: "POST" }),
  },
  labels: {
    list: () => request<Label[]>("/v1/labels"),
    create: (displayName: string) =>
      request<Label>("/v1/labels", { method: "POST", body: JSON.stringify({ displayName }) }),
    delete: (id: string) =>
      request<void>(`/v1/labels/${id}`, { method: "DELETE" }),
  },
}
