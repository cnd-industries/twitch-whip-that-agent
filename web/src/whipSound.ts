// Whip sounds from web/assets, played in turn. New whip-*.mp3 files are
// picked up automatically.
const sounds = Object.entries(
  import.meta.glob<string>('../assets/whip-*.mp3', { eager: true, import: 'default', query: '?url' }),
)
  .sort(([a], [b]) => a.localeCompare(b, undefined, { numeric: true }))
  .map(([, url]) => url)

// Quiet by default so whips do not drown out the stream; the browser source
// URL can override it, e.g. http://127.0.0.1:8080/?volume=0.8
const volume = parseVolume(new URLSearchParams(location.search).get('volume'), 0.3)

let next = 0

export function playWhip() {
  const url = sounds[next]
  next = (next + 1) % sounds.length

  // A fresh element per whip lets rapid whips overlap instead of cutting
  // each other off.
  const audio = new Audio(url)
  audio.volume = volume
  audio.play().catch((err) => {
    console.warn('whip sound:', err)
  })
}

function parseVolume(raw: string | null, fallback: number): number {
  const volume = Number(raw)
  if (raw === null || raw === '' || !Number.isFinite(volume)) {
    return fallback
  }

  return Math.min(Math.max(volume, 0), 1)
}
