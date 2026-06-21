"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Label } from "@/lib/types"
import Link from "next/link"

export default function LabelsPage() {
  const [labels, setLabels] = useState<Label[]>([])
  const [newName, setNewName] = useState("")

  const load = useCallback(async () => {
    try {
      setLabels(await api.labels.list())
    } catch {}
  }, [])

  useEffect(() => { load() }, [load])

  async function create() {
    if (!newName.trim()) return
    await api.labels.create(newName.trim())
    setNewName("")
    load()
  }

  async function remove(id: string) {
    await api.labels.delete(id)
    load()
  }

  return (
    <div className="max-w-md mx-auto">
      <div className="flex items-center gap-4 mb-6">
        <Link href="/" className="text-sm text-gray-500 hover:text-gray-700">Notes</Link>
        <Link href="/archive" className="text-sm text-gray-500 hover:text-gray-700">Archive</Link>
        <Link href="/trash" className="text-sm text-gray-500 hover:text-gray-700">Trash</Link>
        <Link href="/labels" className="text-sm text-blue-600 font-medium">Labels</Link>
      </div>

      <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">Labels</h2>

      <div className="flex gap-2 mb-4">
        <input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && create()}
          placeholder="New label name..."
          className="flex-1 px-3 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:border-blue-400"
        />
        <button onClick={create} className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700">
          Add
        </button>
      </div>

      <div className="space-y-1">
        {labels.map((l) => {
          const id = l.name?.replace("labels/", "") || ""
          return (
            <div key={l.name} className="flex items-center justify-between px-3 py-2 rounded-lg hover:bg-gray-50">
              <span className="text-sm">{l.displayName}</span>
              <button onClick={() => remove(id)} className="text-xs text-red-500 hover:underline">
                Delete
              </button>
            </div>
          )
        })}
        {labels.length === 0 && (
          <p className="text-gray-400 text-sm text-center py-8">No labels yet.</p>
        )}
      </div>
    </div>
  )
}
