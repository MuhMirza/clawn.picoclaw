import { stripBasePath, withBasePath } from "@/lib/base-path"

/** Normalize URL pathname for comparisons (trailing slashes, empty). */
export function normalizePathname(p: string): string {
  const t = p.replace(/\/+$/, "")
  return t === "" ? "/" : t
}

export function isLauncherLoginPathname(pathname: string): boolean {
  return normalizePathname(stripBasePath(pathname)) === "/launcher-login"
}

export function getLauncherLoginPath(): string {
  return withBasePath("/launcher-login")
}
