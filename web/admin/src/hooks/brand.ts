import { useEffect, useState } from 'react'

const FALLBACK = 'blog'

// Reads the public brand name from GET /api/brand (no JWT needed).
// Returns a fallback while loading so the title never flashes empty.
export function useBrand(): string {
  const [brand, setBrand] = useState<string>(FALLBACK)
  useEffect(() => {
    let alive = true
    fetch('/api/brand')
      .then(r => r.ok ? r.json() : null)
      .then(d => { if (alive && d && typeof d.brand === 'string' && d.brand) setBrand(d.brand) })
      .catch(() => {})
    return () => { alive = false }
  }, [])
  return brand
}