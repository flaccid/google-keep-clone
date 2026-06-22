"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Label } from "@/lib/types"
import { Pencil, Trash2, Check, X } from "lucide-react"

export default function LabelsPage() {
  const [labels, setLabels] = useState<Label[]>([])
  const [newName, setNewName] = useState("")
  const [editing, setEditing] = useState<string | null>(null)
  const [editValue, setEditValue] = useState("")

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

  async function rename(id: string) {
    if (!editValue.trim()) return
    await api.labels.update(id, editValue.trim())
    setEditing(null)
    setEditValue("")
    load()
  }

  async function remove(id: string) {
    await api.labels.delete(id)
    load()
  }

  function startEdit(id: string, name: string) {
    setEditing(id)
    setEditValue(name)
  }

  return (
    <div className="max-w-3xl mx-auto px-3 sm:px-6 py-4 sm:py-6">
      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Edit labels</h2>

      <div className="space-y-0.5 mb-6">
        {labels.map((l) => {
          const id = l.name?.replace("labels/", "") || ""
          const isEditing = editing === id
          return (
            <div key={l.name} className="group flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-gray-50 dark:hover:bg-white/5">
              {isEditing ? (
                <>
                  <input
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onKeyDown={(e) => { if (e.key === "Enter") rename(id); if (e.key === "Escape") setEditing(null) }}
                    className="flex-1 px-2 py-0.5 text-sm rounded border border-gray-300 dark:border-[#5f6368] outline-none focus:border-blue-400 bg-white dark:bg-[#2d2e30] text-gray-900 dark:text-[#e8eaed]"
                    autoFocus
                  />
                  <button onClick={() => rename(id)} className="text-green-600 dark:text-green-400 hover:opacity-80"><Check size={16} /></button>
                  <button onClick={() => setEditing(null)} className="text-gray-400 dark:text-[#9aa0a6] hover:opacity-80"><X size={16} /></button>
                </>
              ) : (
                <>
                  <span className="w-5 h-5 rounded-full border border-gray-300 dark:border-[#5f6368] flex items-center justify-center text-[10px] text-gray-400 dark:text-[#9aa0a6] flex-shrink-0">
                    {l.displayName?.charAt(0).toUpperCase()}
                  </span>
                  <span className="flex-1 text-sm text-gray-900 dark:text-[#e8eaed]">{l.displayName}</span>
                  <button onClick={() => startEdit(id, l.displayName || "")} className="sm:opacity-0 sm:group-hover:opacity-100 p-1.5 sm:p-1 text-gray-400 dark:text-[#9aa0a6] hover:text-gray-600 dark:hover:text-[#e8eaed] transition-opacity"><Pencil size={14} /></button>
                  <button onClick={() => remove(id)} className="sm:opacity-0 sm:group-hover:opacity-100 p-1.5 sm:p-1 text-red-400 dark:text-red-500 hover:text-red-600 transition-opacity"><Trash2 size={14} /></button>
                </>
              )}
            </div>
          )
        })}
        {labels.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm text-center py-8">No labels yet.</p>
        )}
      </div>

      <div className="border-t border-gray-100 dark:border-[#3c4043] pt-4">
        <div className="flex items-center gap-2">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && create()}
            placeholder="Create new label..."
            className="flex-1 px-2 py-1 text-sm bg-transparent outline-none placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
          />
          <button onClick={create} className="text-sm text-gray-600 dark:text-[#bdc1c6] hover:text-gray-900 dark:hover:text-[#e8eaed]">
            Done
          </button>
        </div>
      </div>
    </div>
  )
}
