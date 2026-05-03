import profileConfig from './config'
import type {
  LastfmTrack,
  ProfileBadge,
  ProfileConfigSource,
  SocialLink,
} from './schema'

interface ApiConfig {
  baseUrl: string
  lastfmPath: string
  lastfmEnabled: boolean
  viewsPath: string
  refreshIntervalMs: number
}

interface AppConfig {
  content: ProfileConfigSource['content']
  assets: {
    backgroundVideoUrl: string
    bannerUrl: string
    avatarUrl: string
    songUrl: string
  }
  socials: SocialLink[]
  badges: ProfileBadge[]
  features: ProfileConfigSource['features']
  api: ApiConfig
  theme: ProfileConfigSource['theme']
  themeStyles: Record<string, string>
}

const allowedIcons: SocialLink['icon'][] = ['discord', 'github', 'globe', 'youtube']
const fallbackProfileUsername = 'kisakay'

function sanitizePath(path: string) {
  return path.trim().replace(/^\.?\//, '')
}

function sanitizeProfileUsername(username: string) {
  const trimmed = username.trim().toLowerCase()
  if (/^[a-z0-9_-]+$/.test(trimmed)) {
    return trimmed
  }

  return fallbackProfileUsername
}

function sanitizeLastfmUsername(username: string, fallbackUsername: string) {
  const trimmed = username.trim().toLowerCase()
  if (/^[a-z0-9._-]+$/.test(trimmed)) {
    return trimmed
  }

  return sanitizeProfileUsername(fallbackUsername)
}

function appendPathSegment(basePath: string, segment: string) {
  return `${basePath.replace(/\/+$/, '')}/${segment.replace(/^\/+/, '')}`
}

function resolveAssetUrl(path: string) {
  const trimmedPath = path.trim()

  if (/^(?:https?:|data:|blob:)/i.test(trimmedPath)) {
    return trimmedPath
  }

  return `${import.meta.env.BASE_URL}${sanitizePath(trimmedPath)}`
}

function resolveApiBaseUrl(configuredApiBaseUrl: string | undefined) {
  const trimmedBaseUrl = configuredApiBaseUrl?.trim() ?? ''

  if (trimmedBaseUrl) {
    return trimmedBaseUrl.replace(/\/+$/, '')
  }

  if (import.meta.env.DEV) {
    return window.location.origin
  }

  return 'https://profile.kisakay.com'
}

function sanitizeSocialLinks(value: SocialLink[]) {
  const links = value.flatMap((entry) => {
    const label = entry.label.trim()
    const icon = entry.icon.trim() as SocialLink['icon']
    const url = entry.url.trim()

    if (!label || !url || !allowedIcons.includes(icon)) {
      return []
    }

    return [{ label, icon, url }]
  })

  return links
}

function sanitizeBadges(value: ProfileBadge[]) {
  return value.flatMap((entry) => {
    const label = entry.label.trim()
    const icon = entry.icon.trim()

    if (!label || !icon) {
      return []
    }

    return [{ label, icon: resolveAssetUrl(icon) }]
  })
}

function sanitizeColor(value: string, fallback: string) {
  return value.trim() || fallback
}

function sanitizeNumber(value: number, fallback: number) {
  return Number.isFinite(value) && value > 0 ? value : fallback
}

function hexToRgbChannels(color: string) {
  const normalized = color.trim().replace('#', '')
  const hex = normalized.length === 3
    ? normalized.split('').map((char) => `${char}${char}`).join('')
    : normalized

  if (!/^[\da-fA-F]{6}$/.test(hex)) {
    return '255 123 194'
  }

  const red = Number.parseInt(hex.slice(0, 2), 16)
  const green = Number.parseInt(hex.slice(2, 4), 16)
  const blue = Number.parseInt(hex.slice(4, 6), 16)

  return `${red} ${green} ${blue}`
}

const fallbackTheme = {
  accent: '#ff7bc2',
  accentSoft: '#ffb7db',
  accentStrong: '#ff4ba6',
  warning: '#ffd86b',
  warningSoft: '#ffe29a',
}

const theme = {
  accent: sanitizeColor(profileConfig.theme.accent, fallbackTheme.accent),
  accentSoft: sanitizeColor(profileConfig.theme.accentSoft, fallbackTheme.accentSoft),
  accentStrong: sanitizeColor(profileConfig.theme.accentStrong, fallbackTheme.accentStrong),
  warning: sanitizeColor(profileConfig.theme.warning, fallbackTheme.warning),
  warningSoft: sanitizeColor(profileConfig.theme.warningSoft, fallbackTheme.warningSoft),
}

export const appConfig: AppConfig = {
  content: {
    ...profileConfig.content,
    username: sanitizeProfileUsername(profileConfig.content.username),
  },
  assets: {
    backgroundVideoUrl: resolveAssetUrl(profileConfig.assets.backgroundVideoPath),
    bannerUrl: resolveAssetUrl(profileConfig.assets.bannerPath),
    avatarUrl: resolveAssetUrl(profileConfig.assets.avatarPath),
    songUrl: resolveAssetUrl(profileConfig.assets.songPath),
  },
  socials: sanitizeSocialLinks(profileConfig.socials),
  badges: sanitizeBadges(profileConfig.badges),
  features: {
    customCursorEnabled: profileConfig.features.customCursorEnabled !== false,
    cursorHaloEnabled: profileConfig.features.cursorHaloEnabled !== false,
    cardTiltEnabled: profileConfig.features.cardTiltEnabled !== false,
    entryScreenEnabled: profileConfig.features.entryScreenEnabled !== false,
    playerEnabled: profileConfig.features.playerEnabled !== false,
    viewCounterEnabled: profileConfig.features.viewCounterEnabled !== false,
    animatedTitleEnabled: profileConfig.features.animatedTitleEnabled !== false,
    locationEnabled: profileConfig.features.locationEnabled !== false,
    pronounsEnabled: profileConfig.features.pronounsEnabled !== false,
    displayNameEnabled: profileConfig.features.displayNameEnabled !== false,
    bioEnabled: profileConfig.features.bioEnabled !== false,
  },
  api: {
    baseUrl: resolveApiBaseUrl(import.meta.env.VITE_API_BASE_URL),
    lastfmPath: appendPathSegment(
      profileConfig.api.lastfmPath,
      sanitizeLastfmUsername(profileConfig.api.lastfmUsername, profileConfig.content.username),
    ),
    lastfmEnabled: profileConfig.api.lastfmEnabled !== false,
    viewsPath: appendPathSegment(
      profileConfig.api.viewsPath,
      sanitizeProfileUsername(profileConfig.content.username),
    ),
    refreshIntervalMs: sanitizeNumber(profileConfig.api.refreshIntervalMs, 240_000),
  },
  theme,
  themeStyles: {
    '--accent': theme.accent,
    '--accent-soft': theme.accentSoft,
    '--accent-strong': theme.accentStrong,
    '--accent-rgb': hexToRgbChannels(theme.accent),
    '--warning': theme.warning,
    '--warning-soft': theme.warningSoft,
    '--warning-rgb': hexToRgbChannels(theme.warning),
  },
}

export type { LastfmTrack }
