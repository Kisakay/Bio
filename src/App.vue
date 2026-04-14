<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import './App.css'

type IconName = 'discord' | 'github' | 'globe' | 'youtube'
interface SocialLink {
  label: string
  icon: IconName
  url: string
}

interface ProfileBadge {
  label: string
  icon: string
}

interface LastfmTrack {
  title: string
  artist: string
  artwork: string | null
  timestamp: string
  url: string
  isLive: boolean
}

const entered = ref(false)
const backgroundVideoUrl = `${import.meta.env.BASE_URL}assets/background.mp4`
const bannerUrl = `${import.meta.env.BASE_URL}assets/banner.jpg`
const avatarUrl = `${import.meta.env.BASE_URL}assets/pfp.jpg`
const songUrl = `${import.meta.env.BASE_URL}assets/song.mp3`
const apiBaseUrl = resolveApiBaseUrl()
const lastfmTrack = ref<LastfmTrack | null>(null)
const lastfmState = ref<'loading' | 'ready' | 'error'>('loading')
const audioElement = ref<HTMLAudioElement | null>(null)
const audioPlaying = ref(false)
const audioState = ref<'idle' | 'ready' | 'error'>('idle')
const audioCurrentTime = ref(0)
const audioDuration = ref(0)
const viewCount = ref<number | null>(null)
const viewState = ref<'loading' | 'ready' | 'error'>('loading')
const customCursorEnabled = ref(false)
const tiltEnabled = ref(false)
const cursorVisible = ref(false)
const cursorPressed = ref(false)
const cursorInteractive = ref(false)
const cursorX = ref(0)
const cursorY = ref(0)
const cardTiltActive = ref(false)
const cardRotateX = ref(0)
const cardRotateY = ref(0)
const cardShiftX = ref(0)
const cardShiftY = ref(0)
const cardGlowX = ref('50%')
const cardGlowY = ref('24%')

const socialLinks: SocialLink[] = [
  { label: 'Discord', icon: 'discord', url: 'https://discord.com/users/171356978310938624' },
  { label: 'GitHub', icon: 'github', url: 'https://github.com/Kisakay' },
  { label: 'iHorizon', icon: 'globe', url: 'https://www.ihorizon.org' },
  { label: 'YouTube', icon: 'youtube', url: 'https://youtube.com/@Kisakay' },
]

const profileBadges: ProfileBadge[] = [
  { label: 'Bug Hunter Gold', icon: `${import.meta.env.BASE_URL}assets/discord_badges/bug_hunter_level_2.svg` },
  { label: 'Early Verified Bot Developer', icon: `${import.meta.env.BASE_URL}assets/discord_badges/early_verified_developer.svg` },
  { label: 'Server Booster 24 Months', icon: `${import.meta.env.BASE_URL}assets/discord_badges/boosting_24_months.svg` },
]

let refreshTimer: number | undefined
let titleAnimationTimer: number | undefined
let removeCursorListeners: (() => void) | undefined

onMounted(() => {
  const supportsFinePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches

  entered.value = false
  tiltEnabled.value = true;
  startTitleAnimation()
  void updateNowPlaying()
  void registerView()

  refreshTimer = window.setInterval(() => {
    void updateNowPlaying()
  }, 60_000)

  if (supportsFinePointer) {
    customCursorEnabled.value = true
    removeCursorListeners = attachCursorListeners()
  }
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }

  if (titleAnimationTimer) {
    window.clearInterval(titleAnimationTimer)
  }

  removeCursorListeners?.()
})

function enterProfile() {
  entered.value = true
  void playAudio()
}

function resolveApiBaseUrl() {
  const configuredApiBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim()

  if (configuredApiBaseUrl) {
    return configuredApiBaseUrl.replace(/\/+$/, '')
  }

  if (import.meta.env.DEV) {
    return window.location.origin
  }

  return 'https://profile.kisakay.com'
}

function buildApiUrl(path: string) {
  return new URL(path, `${apiBaseUrl}/`).toString()
}

function startTitleAnimation() {
  const title = "@2h0 | Kisakay";
  const frames = Array.from(
    { length: title.length },
    (_, i) => title.slice(0, title.length - i)
  );
  let frameIndex = 0
  let direction = 1

  document.title = frames[frameIndex]

  titleAnimationTimer = window.setInterval(() => {
    if (frameIndex === frames.length - 1) {
      direction = -1
    } else if (frameIndex === 0) {
      direction = 1
    }

    frameIndex += direction
    document.title = frames[frameIndex]
  }, 200)
}

async function updateNowPlaying() {
  if (!lastfmTrack.value) {
    lastfmState.value = 'loading'
  }

  try {
    const response = await fetch(buildApiUrl('/api/lastfm'), {
      cache: 'no-store',
    })

    if (!response.ok) {
      throw new Error(`Last.fm proxy request failed with ${response.status}`)
    }

    const payload = (await response.json()) as LastfmTrack | null

    if (!payload) {
      throw new Error('Unable to parse Last.fm scrobble')
    }

    lastfmTrack.value = payload
    lastfmState.value = 'ready'
  } catch (error) {
    console.error(error)
    lastfmState.value = 'error'
  }
}

async function registerView() {
  viewState.value = 'loading'

  try {
    const response = await fetch(buildApiUrl('/api/views'), {
      method: 'POST',
      headers: {
        Accept: 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      throw new Error(`View counter request failed with ${response.status}`)
    }

    const payload = (await response.json()) as { count?: unknown }

    if (typeof payload.count !== 'number') {
      throw new Error('Invalid view counter payload')
    }

    viewCount.value = payload.count
    viewState.value = 'ready'
  } catch (error) {
    console.error(error)
    void fetchViewCount()
  }
}

async function fetchViewCount() {
  try {
    const response = await fetch(buildApiUrl('/api/views'), {
      cache: 'no-store',
    })

    if (!response.ok) {
      throw new Error(`View counter fallback failed with ${response.status}`)
    }

    const payload = (await response.json()) as { count?: unknown }

    if (typeof payload.count !== 'number') {
      throw new Error('Invalid fallback view counter payload')
    }

    viewCount.value = payload.count
    viewState.value = 'ready'
  } catch (error) {
    console.error(error)
    viewState.value = 'error'
  }
}

async function playAudio() {
  if (!audioElement.value) {
    return
  }

  try {
    await audioElement.value.play()
  } catch (error) {
    console.error(error)
  }
}

function toggleAudioPlayback() {
  if (!audioElement.value) {
    return
  }

  if (audioElement.value.paused) {
    void playAudio()
    return
  }

  audioElement.value.pause()
}

function handleAudioReady() {
  audioState.value = 'ready'
}

function handleAudioPlay() {
  audioPlaying.value = true
}

function handleAudioPause() {
  audioPlaying.value = false
}

function handleAudioError() {
  audioState.value = 'error'
}

function handleAudioTimeUpdate() {
  if (!audioElement.value) {
    return
  }

  audioCurrentTime.value = audioElement.value.currentTime
}

function handleAudioMetadata() {
  if (!audioElement.value) {
    return
  }

  audioDuration.value = Number.isFinite(audioElement.value.duration)
    ? audioElement.value.duration
    : 0
}

function getAudioProgress() {
  if (!audioDuration.value || audioState.value === 'error') {
    return '0%'
  }

  const progress = Math.min(audioCurrentTime.value / audioDuration.value, 1)
  return `${progress * 100}%`
}

function formatViewCount(count: number | null) {
  if (count === null) {
    return viewState.value === 'error' ? 'offline' : '...'
  }

  return new Intl.NumberFormat('en-US').format(count)
}

function attachCursorListeners() {
  const handleMouseMove = (event: MouseEvent) => {
    cursorX.value = event.clientX
    cursorY.value = event.clientY
    cursorVisible.value = true
    cursorInteractive.value = isInteractiveTarget(event.target)
  }

  const handlePointerDown = () => {
    cursorPressed.value = true
  }

  const handlePointerUp = (event: PointerEvent) => {
    cursorPressed.value = false
    cursorInteractive.value = isInteractiveTarget(event.target)
  }

  const handlePointerLeave = () => {
    cursorVisible.value = false
    cursorPressed.value = false
    cursorInteractive.value = false
  }

  window.addEventListener('mousemove', handleMouseMove)
  window.addEventListener('pointerdown', handlePointerDown)
  window.addEventListener('pointerup', handlePointerUp)
  window.addEventListener('blur', handlePointerLeave)
  document.addEventListener('mouseleave', handlePointerLeave)

  return () => {
    window.removeEventListener('mousemove', handleMouseMove)
    window.removeEventListener('pointerdown', handlePointerDown)
    window.removeEventListener('pointerup', handlePointerUp)
    window.removeEventListener('blur', handlePointerLeave)
    document.removeEventListener('mouseleave', handlePointerLeave)
  }
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof Element && !!target.closest('a, button, [role="button"]')
}

function handleCardPointerMove(event: PointerEvent) {
  if (!tiltEnabled.value) {
    return
  }

  const target = event.currentTarget

  if (!(target instanceof HTMLElement)) {
    return
  }

  const rect = target.getBoundingClientRect()
  const relativeX = (event.clientX - rect.left) / rect.width
  const relativeY = (event.clientY - rect.top) / rect.height
  const offsetX = relativeX - 0.5
  const offsetY = relativeY - 0.5
  const edgeBoostX = Math.sign(offsetX) * Math.pow(Math.abs(offsetX) * 2, 1.8)
  const edgeBoostY = Math.sign(offsetY) * Math.pow(Math.abs(offsetY) * 2, 1.8)

  cardTiltActive.value = true
  cardRotateY.value = edgeBoostX * 24
  cardRotateX.value = edgeBoostY * -22
  cardShiftX.value = edgeBoostX * 20
  cardShiftY.value = edgeBoostY * 18
  cardGlowX.value = `${relativeX * 100}%`
  cardGlowY.value = `${relativeY * 100}%`
}

function resetCardTilt() {
  cardTiltActive.value = false
  cardRotateX.value = 0
  cardRotateY.value = 0
  cardShiftX.value = 0
  cardShiftY.value = 0
  cardGlowX.value = '50%'
  cardGlowY.value = '24%'
}
</script>

<template>
  <div class="scene" :class="{ 'scene--custom-cursor': customCursorEnabled }">
    <audio ref="audioElement" :src="songUrl" loop preload="auto" @canplay="handleAudioReady" @error="handleAudioError"
      @loadedmetadata="handleAudioMetadata" @play="handleAudioPlay" @pause="handleAudioPause"
      @timeupdate="handleAudioTimeUpdate"></audio>

    <video class="scene__video" :src="backgroundVideoUrl" autoplay muted loop playsinline preload="auto"
      aria-hidden="true"></video>
    <div class="scene__wallpaper"></div>
    <div class="scene__vignette"></div>
    <div class="scene__grain"></div>

    <button v-if="!entered" class="entry" type="button" aria-label="Enter profile" @click="enterProfile">
      <span class="entry__label">click here</span>
    </button>

    <div class="corner-player" :class="{ 'corner-player--visible': entered }">
      <button class="corner-player__button" type="button" :aria-pressed="audioPlaying"
        :aria-label="audioPlaying ? 'Pause audio' : 'Play audio'" @click="toggleAudioPlayback">
        <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path v-if="audioPlaying"
            d="M8 6.75C8 5.78 8.78 5 9.75 5s1.75.78 1.75 1.75v10.5C11.5 18.22 10.72 19 9.75 19S8 18.22 8 17.25V6.75Zm4.5 0C12.5 5.78 13.28 5 14.25 5S16 5.78 16 6.75v10.5c0 .97-.78 1.75-1.75 1.75s-1.75-.78-1.75-1.75V6.75Z"
            fill="currentColor" />
          <path v-else
            d="M9.18 6.58c0-1.07 1.17-1.73 2.08-1.18l6.67 4.06c.88.54.88 1.82 0 2.36l-6.67 4.06c-.91.55-2.08-.11-2.08-1.18V6.58Z"
            fill="currentColor" />
        </svg>
      </button>
      <div class="corner-player__copy">
        <strong>song.mp3</strong>
        <div class="corner-player__progress" aria-hidden="true">
          <span class="corner-player__progress-fill" :style="{ width: getAudioProgress() }"></span>
        </div>
      </div>
    </div>

    <main class="profile" :class="{ 'profile--visible': entered }">
      <section class="profile-card" :class="{ 'profile-card--tilting': cardTiltActive && tiltEnabled }" :style="{
        '--card-rotate-x': `${cardRotateX}deg`,
        '--card-rotate-y': `${cardRotateY}deg`,
        '--card-shift-x': `${cardShiftX}px`,
        '--card-shift-y': `${cardShiftY}px`,
        '--card-glow-x': cardGlowX,
        '--card-glow-y': cardGlowY,
      }" @pointermove="handleCardPointerMove" @pointerleave="resetCardTilt" @pointercancel="resetCardTilt">
        <div class="profile-card__banner">
          <img class="profile-card__banner-image" :src="bannerUrl" alt="Kisakay banner" />
          <div class="profile-card__banner-shade"></div>
          <div class="profile-card__banner-grid"></div>
          <div class="profile-card__banner-copy">
            <span>kisakay.com</span>
          </div>
        </div>

        <div class="profile-card__body">
          <div class="avatar-ring">
            <img :src="avatarUrl" alt="Portrait of Kisakay" />
          </div>

          <div class="identity">
            <div class="identity__heading">
              <h1>@2h0</h1>
              <div class="identity__badges" aria-label="Discord badges">
                <span v-for="badge in profileBadges" :key="badge.label" class="identity__badge" :title="badge.label"
                  :aria-label="badge.label">
                  <img :src="badge.icon" :alt="badge.label" />
                </span>
              </div>
            </div>
            <p class="identity__subtitle">Kisakay</p>
            <p class="identity__pronouns">she / her</p>
            <div class="identity__location" aria-label="Location">
              <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path d="M12 21s6-4.35 6-10a6 6 0 1 0-12 0c0 5.65 6 10 6 10Z" stroke="currentColor"
                  stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" />
                <circle cx="12" cy="11" r="2.25" stroke="currentColor" stroke-width="1.5" />
              </svg>
              <span>Somewhere in Brittany</span>
            </div>
            <p class="identity__bio">
              developer, internet enjoyer.
            </p>
          </div>

          <nav class="social-row" aria-label="Social links">
            <a v-for="link in socialLinks" :key="link.label" :href="link.url" class="social-row__link"
              :aria-label="link.label" target="_blank" rel="noreferrer">
              <svg class="social-row__icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path v-if="link.icon === 'discord'"
                  d="M20.317 4.37a19.79 19.79 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.212.375-.444.864-.608 1.249a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.078.078 0 0 0-.079-.036A19.74 19.74 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.058a.082.082 0 0 0 .031.056 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.1 14.1 0 0 0 1.226-1.99.077.077 0 0 0-.041-.106 13 13 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .078-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .079.01c.12.099.245.196.372.292a.077.077 0 0 1-.006.128 12.3 12.3 0 0 1-1.873.892.077.077 0 0 0-.04.107c.36.698.772 1.362 1.225 1.989a.076.076 0 0 0 .084.028 19.85 19.85 0 0 0 6.002-3.03.077.077 0 0 0 .032-.055c.5-5.177-.838-9.669-3.549-13.66a.061.061 0 0 0-.031-.028ZM8.02 15.331c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.418 2.157-2.418 1.21 0 2.175 1.094 2.157 2.418 0 1.334-.956 2.419-2.157 2.419Zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.418 2.157-2.418 1.21 0 2.175 1.094 2.157 2.418 0 1.334-.947 2.419-2.157 2.419Z"
                  fill="currentColor" />
                <path v-if="link.icon === 'github'"
                  d="M9 18.4c-3.75 1.12-3.75-1.88-5.25-2.25m10.5 4.5v-2.9c.03-.38-.06-.76-.25-1.1-.2-.34-.48-.63-.82-.84 2.78-.31 5.7-1.36 5.7-6.16a4.82 4.82 0 0 0-1.3-3.34 4.45 4.45 0 0 0-.08-3.3s-1.05-.31-3.45 1.28a11.8 11.8 0 0 0-6.3 0C5.35 2.7 4.3 3 4.3 3a4.45 4.45 0 0 0-.08 3.3 4.82 4.82 0 0 0-1.3 3.34c0 4.77 2.9 5.85 5.68 6.17-.34.2-.63.49-.82.84-.19.34-.28.72-.25 1.1v2.9"
                  stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.45" />
                <path v-if="link.icon === 'globe'"
                  d="M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Zm0 0c2.33-2.46 3.66-5.72 3.75-9 0-3.28-1.42-6.54-3.75-9-2.33 2.46-3.66 5.72-3.75 9 .09 3.28 1.42 6.54 3.75 9ZM4 12h16"
                  stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.45" />
                <path v-if="link.icon === 'youtube'"
                  d="M16.76 7.24c-2.78-.34-6.75-.34-9.52 0A2.7 2.7 0 0 0 4.9 9.48c-.2 1.68-.2 3.37 0 5.04a2.7 2.7 0 0 0 2.34 2.24c2.77.34 6.74.34 9.52 0a2.7 2.7 0 0 0 2.34-2.24c.2-1.67.2-3.36 0-5.04a2.7 2.7 0 0 0-2.34-2.24Z"
                  stroke="currentColor" stroke-linejoin="round" stroke-width="1.45" />
                <path v-if="link.icon === 'youtube'" d="m10.35 9.56 4.2 2.44-4.2 2.44V9.56Z" fill="currentColor" />
              </svg>
              <span class="sr-only">{{ link.label }}</span>
            </a>
          </nav>

          <div class="activity">
            <div class="activity__meta">
              <span>#1 nowplaying</span>
              <span>{{
                lastfmState === 'error'
                  ? 'service status'
                  : lastfmTrack?.timestamp ?? 'loading'
              }}</span>
            </div>
            <div class="activity__content" :class="{ 'activity__content--offline': lastfmState === 'error' }">
              <img v-if="lastfmTrack?.artwork && lastfmState !== 'error'" class="activity__art"
                :src="lastfmTrack.artwork" alt="" />
              <div v-else-if="lastfmState === 'error'" class="activity__offline-orb" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M12 3.75 20.25 18a1.5 1.5 0 0 1-1.3 2.25H5.05A1.5 1.5 0 0 1 3.75 18L12 3.75Z"
                    stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" />
                  <path d="M12 9v4.5" stroke="currentColor" stroke-linecap="round" stroke-width="1.8" />
                  <circle cx="12" cy="16.5" r="1" fill="currentColor" />
                </svg>
              </div>
              <div class="activity__copy" :class="{ 'activity__copy--offline': lastfmState === 'error' }">
                <strong :class="{ 'activity__title--offline': lastfmState === 'error' }">
                  {{
                    lastfmTrack?.title ??
                    (lastfmState === 'error'
                      ? 'Now playing unavailable'
                      : 'loading last.fm scrobble')
                  }}
                </strong>
                <a v-if="lastfmTrack && lastfmState !== 'error'" class="activity__artist" :href="lastfmTrack.url"
                  target="_blank" rel="noreferrer">
                  {{ lastfmTrack.artist }}
                </a>
                <div v-else-if="lastfmState === 'error'" class="activity__status activity__status--offline">
                  <span class="activity__status-label">API OFFLINE</span>
                </div>
                <span v-else class="activity__artist">
                  Last.fm / Kisakay
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <div class="corner-views" :class="{ 'corner-views--visible': entered }" aria-live="polite">
      <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M2.25 12s3.7-6.75 9.75-6.75S21.75 12 21.75 12s-3.7 6.75-9.75 6.75S2.25 12 2.25 12Z"
          stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.45" />
        <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.45" />
      </svg>
      <div class="corner-views__copy">
        <strong>{{ formatViewCount(viewCount) }}</strong>
      </div>
    </div>

    <div v-if="customCursorEnabled" class="cursor-layer" :class="{
      'cursor-layer--visible': cursorVisible,
      'cursor-layer--interactive': cursorInteractive,
      'cursor-layer--pressed': cursorPressed,
    }" :style="{ transform: `translate3d(${cursorX}px, ${cursorY}px, 0)` }" aria-hidden="true">
      <div class="cursor-layer__halo"></div>
      <div class="cursor-layer__glow"></div>
    </div>
  </div>
</template>
