export type IconName = 'discord' | 'github' | 'globe' | 'youtube'

export interface SocialLink {
  label: string
  icon: IconName
  url: string
}

export interface ProfileBadge {
  label: string
  icon: string
}

export interface LastfmTrack {
  title: string
  artist: string
  artwork: string | null
  timestamp: string
  url: string
  isLive: boolean
}

export interface ProfileContent {
  username: string
  siteTitle: string
  enterButtonLabel: string
  enterButtonAriaLabel: string
  audioPlayLabel: string
  audioPauseLabel: string
  playerTrackLabel: string
  bannerLabel: string
  bannerAlt: string
  avatarAlt: string
  handle: string
  displayName: string
  pronouns: string
  badgesAriaLabel: string
  locationAriaLabel: string
  location: string
  bio: string
  socialNavAriaLabel: string
  nowPlayingLabel: string
  lastfmLoadingTimestampLabel: string
  lastfmServiceStatusLabel: string
  lastfmUnavailableTitle: string
  lastfmLoadingTitle: string
  lastfmOfflineLabel: string
  lastfmFallbackArtist: string
  viewsOfflineLabel: string
  viewsLoadingLabel: string
  numberLocale: string
}

export interface ProfileAssets {
  backgroundVideoPath: string
  bannerPath: string
  avatarPath: string
  songPath: string
}

export interface ProfileTheme {
  accent: string
  accentSoft: string
  accentStrong: string
  warning: string
  warningSoft: string
}

export interface ProfileFeatures {
  customCursorEnabled: boolean
  cursorHaloEnabled: boolean
  cardTiltEnabled: boolean
  entryScreenEnabled: boolean
  playerEnabled: boolean
  viewCounterEnabled: boolean
  animatedTitleEnabled: boolean
}

export interface StaticApiConfig {
  lastfmPath: string
  lastfmUsername: string
  lastfmEnabled: boolean
  viewsPath: string
  refreshIntervalMs: number
}

export interface ProfileConfigSource {
  content: ProfileContent
  assets: ProfileAssets
  socials: SocialLink[]
  badges: ProfileBadge[]
  features: ProfileFeatures
  api: StaticApiConfig
  theme: ProfileTheme
}
