function normalizeBasePath(input: string | null | undefined): string {
  if (!input) return ""
  let value = input.trim()
  if (!value || value === "/") return ""
  if (!value.startsWith("/")) value = `/${value}`
  value = value.replace(/\/+$/g, "")
  return value === "/" ? "" : value
}

declare global {
  interface Window {
    __PICO_BASE_PATH__?: string
  }
}

export function getBasePath(): string {
  if (typeof globalThis === "undefined") return ""

  const win = globalThis as typeof globalThis & { __PICO_BASE_PATH__?: string }
  const explicit = normalizeBasePath(win.__PICO_BASE_PATH__)
  if (explicit) return explicit

  const pathname = globalThis.location?.pathname || "/"
  const match = pathname.match(/^\/launcher\/[^/]+/)
  if (match) return normalizeBasePath(match[0])

  return ""
}

export function withBasePath(path: string): string {
  const normalizedBase = getBasePath()
  const normalizedPath = path.startsWith("/") ? path : `/${path}`

  if (!normalizedBase) return normalizedPath
  if (normalizedPath === "/") return normalizedBase || "/"
  return `${normalizedBase}${normalizedPath}`
}

export function stripBasePath(pathname: string): string {
  const normalizedPathname = pathname || "/"
  const base = getBasePath()
  if (!base) return normalizedPathname
  if (normalizedPathname === base) return "/"
  if (normalizedPathname.startsWith(`${base}/`)) {
    return normalizedPathname.slice(base.length) || "/"
  }
  return normalizedPathname
}

export function buildApiPath(path: string): string {
  return withBasePath(path)
}

export function getProjectIdFromBasePath(): string {
  const base = getBasePath()
  const match = base.match(/^\/launcher\/([^/]+)$/)
  return match?.[1] || ""
}
