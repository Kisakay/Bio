:root {
  --bg: #050505;
  --panel: rgba(10, 10, 10, 0.56);
  --panel-strong: rgba(16, 16, 18, 0.86);
  --line: rgba(255, 255, 255, 0.11);
  --text: #f5f3ef;
  --muted: rgba(245, 243, 239, 0.7);
  --shadow: 0 30px 80px rgba(0, 0, 0, 0.5);
  --accent: #ff7bc2;
  --accent-soft: #ffb7db;
  --accent-strong: #ff4ba6;
  --accent-rgb: 255 123 194;
  --warning: #ffd86b;
  --warning-soft: #ffe29a;
  --warning-rgb: 255 216 107;
}

.scene {
  position: relative;
  display: grid;
  place-items: center;
  min-height: 100vh;
  overflow: hidden;
  padding: 1rem;
  background: #040404;
}

.scene__video,
.scene__wallpaper,
.scene__vignette,
.scene__grain {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.scene__video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scale(1.05);
  filter: saturate(1.05) brightness(0.72);
}

.scene__wallpaper {
  background:
    radial-gradient(circle at 22% 18%, rgba(255, 255, 255, 0.08), transparent 24%),
    linear-gradient(rgba(6, 6, 7, 0.24), rgba(6, 6, 7, 0.58));
  backdrop-filter: blur(6px);
}

.scene__vignette {
  background:
    radial-gradient(circle at center, transparent 22%, rgba(0, 0, 0, 0.38) 70%),
    linear-gradient(180deg, rgba(0, 0, 0, 0.42), rgba(0, 0, 0, 0.55));
}

.scene__grain {
  opacity: 0.18;
  background-image:
    radial-gradient(circle at 25% 20%, rgba(255, 255, 255, 0.08) 0 1px, transparent 1px),
    radial-gradient(circle at 70% 35%, rgba(255, 255, 255, 0.05) 0 1px, transparent 1px),
    radial-gradient(circle at 40% 75%, rgba(255, 255, 255, 0.06) 0 1px, transparent 1px);
  background-size: 160px 160px;
}

.entry {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: grid;
  place-items: center;
  border: 0;
  overflow: hidden;
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0.08), transparent 26%),
    linear-gradient(180deg, rgba(7, 7, 10, 0.18), rgba(7, 7, 10, 0.34));
  color: var(--text);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.9rem;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  backdrop-filter: blur(18px) saturate(1.15);
  -webkit-backdrop-filter: blur(18px) saturate(1.15);
  cursor: pointer;
  transition: background 180ms ease, backdrop-filter 180ms ease;
}

.entry::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px);
  background-size: 34px 34px;
  opacity: 0.3;
  pointer-events: none;
}

.entry__label {
  position: relative;
  z-index: 1;
  padding: 1rem 1.35rem;
  border: 1px solid rgb(var(--accent-rgb) / 0.22);
  border-radius: 999px;
  background: rgb(var(--accent-rgb) / 0.08);
  color: rgba(245, 243, 239, 0.82);
  box-shadow:
    0 12px 36px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgb(var(--accent-rgb) / 0.18);
  transition:
    color 180ms ease,
    background 180ms ease,
    border-color 180ms ease,
    box-shadow 180ms ease,
    text-shadow 180ms ease,
    transform 180ms ease;
}

.entry:hover {
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0.12), transparent 28%),
    linear-gradient(180deg, rgba(7, 7, 10, 0.12), rgba(7, 7, 10, 0.28));
}

.entry:hover .entry__label {
  transform: translateY(-1px);
  color: #ffffff;
  background: linear-gradient(180deg, rgb(var(--accent-rgb) / 0.18), rgb(var(--accent-rgb) / 0.08));
  border-color: rgb(var(--accent-rgb) / 0.42);
  text-shadow: 0 0 18px rgb(var(--accent-rgb) / 0.42);
  box-shadow:
    0 18px 42px rgba(0, 0, 0, 0.22),
    0 0 28px rgb(var(--accent-rgb) / 0.16),
    inset 0 1px 0 rgb(var(--accent-rgb) / 0.24);
}

.corner-player,
.corner-views {
  position: fixed;
  z-index: 22;
  display: inline-flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.8rem 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.11);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03)),
    rgba(8, 8, 10, 0.36);
  box-shadow:
    0 20px 52px rgba(0, 0, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(18px) saturate(1.1);
  -webkit-backdrop-filter: blur(18px) saturate(1.1);
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 180ms ease,
    transform 180ms ease,
    border-color 180ms ease,
    background 180ms ease;
}

.corner-player--visible,
.corner-views--visible {
  opacity: 1;
  pointer-events: auto;
}

.corner-player {
  top: 1rem;
  left: 1rem;
  min-width: 220px;
  transform: translate3d(-12px, -8px, 0);
}

.corner-player__button {
  flex: none;
  width: 2.85rem;
  height: 2.85rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text);
  box-shadow:
    0 12px 26px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
  transition:
    transform 180ms ease,
    border-color 180ms ease,
    background 180ms ease,
    box-shadow 180ms ease;
}

.corner-player__button:hover {
  transform: translateY(-1px);
  border-color: rgb(var(--accent-rgb) / 0.3);
  background: rgb(var(--accent-rgb) / 0.12);
  box-shadow:
    0 16px 32px rgba(0, 0, 0, 0.22),
    0 0 18px rgb(var(--accent-rgb) / 0.12),
    inset 0 1px 0 rgb(var(--accent-rgb) / 0.16);
}

.corner-player__button svg {
  width: 1.15rem;
  height: 1.15rem;
}

.corner-player__copy,
.corner-views__copy {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
}

.corner-player__copy span,
.corner-views__copy span {
  font-family: 'JetBrains Mono', monospace;
  text-transform: uppercase;
}

.corner-player__copy span,
.corner-views__copy span {
  color: rgba(245, 243, 239, 0.58);
  font-size: 0.6rem;
  letter-spacing: 0.12em;
}

.corner-player__copy strong,
.corner-views__copy strong {
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.corner-player__progress {
  position: relative;
  width: 120px;
  height: 4px;
  margin-top: 0.22rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
}

.corner-player__progress-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--accent-strong), var(--accent-soft));
  box-shadow: 0 0 14px rgb(var(--accent-rgb) / 0.28);
  transition: width 120ms linear;
}

.corner-views {
  z-index: 22;
  right: 1rem;
  bottom: 1rem;
  transform: translate3d(12px, 8px, 0);
}

.corner-views svg {
  width: 1.2rem;
  height: 1.2rem;
  color: rgba(255, 255, 255, 0.9);
  flex: none;
}

.corner-views__copy {
  display: grid;
  gap: 0.18rem;
}

.corner-views__copy span {
  font-size: 0.6rem;
  letter-spacing: 0.12em;
}

.corner-views__copy strong {
  font-size: 0.9rem;
  font-weight: 600;
}

.profile {
  position: relative;
  z-index: 2;
  display: grid;
  place-items: center;
  width: 100%;
  perspective: 1800px;
  opacity: 0;
  transform: translateY(16px) scale(0.985);
  transition: opacity 220ms ease, transform 220ms ease;
}

.profile--visible {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.profile-card {
  --card-rotate-x: 0deg;
  --card-rotate-y: 0deg;
  --card-shift-x: 0px;
  --card-shift-y: 0px;
  --card-glow-x: 50%;
  --card-glow-y: 24%;
  width: min(100%, 380px);
  min-height: 650px;
  position: relative;
  margin-inline: auto;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 26px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03)),
    var(--panel);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.035) inset,
    0 1px 0 rgba(255, 255, 255, 0.06) inset,
    var(--shadow);
  backdrop-filter: blur(16px);
  transform:
    translate3d(var(--card-shift-x), var(--card-shift-y), 0)
    rotateX(var(--card-rotate-x))
    rotateY(var(--card-rotate-y));
  transform-style: preserve-3d;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    border-color 180ms ease;
}

.profile-card--compact {
  min-height: 590px;
}

.profile-card--sparse {
  width: min(100%, 380px);
  min-height: 0;
}

.profile-card--minimal {
  width: min(100%, 360px);
  min-height: 0;
}

.profile-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  border: 1px solid rgba(255, 255, 255, 0.035);
  pointer-events: none;
}

.profile-card::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background:
    radial-gradient(circle at var(--card-glow-x) var(--card-glow-y), rgb(var(--accent-rgb) / 0.28), transparent 28%),
    linear-gradient(135deg, rgb(var(--accent-rgb) / 0.14), transparent 34%);
  opacity: 0.5;
  mix-blend-mode: screen;
  pointer-events: none;
  transition: opacity 180ms ease;
}

.profile-card--tilting {
  border-color: rgb(var(--accent-rgb) / 0.28);
  box-shadow:
    0 0 0 1px rgb(var(--accent-rgb) / 0.08) inset,
    0 1px 0 rgb(var(--accent-rgb) / 0.16) inset,
    0 52px 140px rgba(0, 0, 0, 0.62),
    0 0 40px rgb(var(--accent-rgb) / 0.12);
}

.profile-card--tilting::after {
  opacity: 0.8;
}

.cursor-layer {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 60;
  opacity: 0;
  pointer-events: none;
  will-change: transform, opacity;
  transition:
    opacity 120ms ease,
    transform 50ms linear;
}

.cursor-layer--visible {
  opacity: 1;
}

.cursor-layer__halo,
.cursor-layer__glow {
  position: absolute;
  top: 0;
  left: 0;
}

.cursor-layer__halo {
  width: 136px;
  height: 136px;
  border-radius: 999px;
  background: radial-gradient(circle, rgb(var(--accent-rgb) / 0.5) 0%, rgb(var(--accent-rgb) / 0.18) 30%, rgb(var(--accent-rgb) / 0.08) 54%, transparent 76%);
  filter: blur(14px);
  transform: translate(-66px, -66px);
  mix-blend-mode: screen;
}

.cursor-layer__glow {
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: radial-gradient(circle, var(--accent-soft), rgb(var(--accent-rgb) / 0.24) 60%, transparent 78%);
  filter: blur(7px);
  transform: translate(-8px, -8px);
  opacity: 1;
}

.cursor-layer--interactive .cursor-layer__halo {
  width: 168px;
  height: 168px;
  transform: translate(-82px, -82px);
  background: radial-gradient(circle, rgb(var(--accent-rgb) / 0.62) 0%, rgb(var(--accent-rgb) / 0.22) 32%, rgb(var(--accent-rgb) / 0.08) 56%, transparent 76%);
}

.cursor-layer--interactive .cursor-layer__glow {
  opacity: 1;
  filter: blur(8px);
}

.cursor-layer--pressed .cursor-layer__halo {
  transform: translate(-66px, -66px) scale(0.9);
}

.profile-card__banner {
  position: relative;
  height: 188px;
  overflow: hidden;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.profile-card--minimal .profile-card__banner {
  height: 172px;
}

.profile-card__banner::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.08), transparent 28%);
}

.profile-card__banner-grid,
.profile-card__banner-image,
.profile-card__banner-shade,
.profile-card__banner-copy {
  position: absolute;
}

.profile-card__banner-image {
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 42%;
}

.profile-card__banner-shade {
  inset: 0;
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0.06), transparent 32%),
    linear-gradient(180deg, transparent 36%, rgba(8, 8, 10, 0.82));
}

.profile-card__banner-grid {
  inset: 0;
  background:
    linear-gradient(rgba(255, 255, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.045) 1px, transparent 1px);
  background-size: 28px 28px;
  opacity: 0.14;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.68), transparent 88%);
}

.profile-card__banner-copy {
  inset: auto auto 1rem 1rem;
  z-index: 2;
  display: grid;
  gap: 0.25rem;
  text-shadow: 0 4px 18px rgba(0, 0, 0, 0.35);
}

.profile-card__banner-copy span,
.profile-card__banner-copy strong {
  text-transform: uppercase;
}

.profile-card__banner-copy span {
  color: rgba(245, 243, 239, 0.58);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.62rem;
  letter-spacing: 0.16em;
}

.profile-card__banner-copy strong {
  color: rgba(255, 255, 255, 0.96);
  font-family: 'Syne', sans-serif;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.profile-card__body {
  display: grid;
  justify-items: center;
  gap: 1rem;
  padding: 0 1.2rem 1.25rem;
}

.profile-card__body--sparse {
  gap: 0.85rem;
  padding-bottom: 1.05rem;
}

.profile-card__body--minimal {
  gap: 0.7rem;
  padding-bottom: 0.9rem;
}

.avatar-ring {
  width: 96px;
  height: 96px;
  margin-top: -3rem;
  padding: 4px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
  box-shadow: 0 14px 38px rgba(0, 0, 0, 0.4);
  position: relative;
  z-index: 2;
}

.avatar-ring--minimal {
  margin-top: -2.7rem;
}

.avatar-ring img {
  width: 100%;
  height: 100%;
  border-radius: 999px;
  object-fit: cover;
}

.identity {
  text-align: center;
}

.identity--minimal {
  display: grid;
  gap: 0.2rem;
}

.identity__heading {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.identity h1 {
  margin: 0;
  font-family: 'Syne', sans-serif;
  font-size: clamp(2.6rem, 9vw, 3.4rem);
  line-height: 0.9;
}

.identity__badges {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  padding: 0.36rem 0.48rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.055);
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.identity__badge {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.45rem;
  height: 1.45rem;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease;
}

.identity__badge img {
  width: 1.15rem;
  height: 1.15rem;
  display: block;
  object-fit: contain;
  transition:
    transform 180ms ease,
    filter 180ms ease;
}

.identity__badge::after {
  content: attr(aria-label);
  position: absolute;
  left: 50%;
  bottom: calc(100% + 0.55rem);
  transform: translateX(-50%) translateY(4px);
  padding: 0.42rem 0.56rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 10px;
  background: rgba(12, 12, 16, 0.92);
  color: rgba(255, 255, 255, 0.92);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.58rem;
  letter-spacing: 0.08em;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.24);
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.identity__badge:hover {
  transform: translateY(-1px);
}

.identity__badge:hover img {
  transform: scale(1.06);
  filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.18));
}

.identity__badge:hover::after {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

.identity__subtitle,
.identity__pronouns,
.identity__bio,
.activity span,
.stats span,
.badge-row span,
.entry {
  font-family: 'JetBrains Mono', monospace;
}

.identity__subtitle {
  margin: 0.45rem 0 0;
  color: var(--text);
  font-size: 0.92rem;
}

.identity__pronouns {
  margin: 0.35rem 0 0;
  color: var(--muted);
  font-size: 0.68rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.identity__location {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 0.7rem;
  padding: 0.7rem 0.95rem;
  border: 1px solid rgb(var(--accent-rgb) / 0.24);
  border-radius: 999px;
  background: linear-gradient(180deg, rgb(var(--accent-rgb) / 0.08), rgb(var(--accent-rgb) / 0.04));
  color: rgba(245, 243, 239, 0.88);
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.12),
    inset 0 1px 0 rgb(var(--accent-rgb) / 0.12);
}

.identity__location svg {
  width: 0.95rem;
  height: 0.95rem;
  color: var(--accent-soft);
}

.identity__location span {
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.68rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.identity__bio {
  margin: 0.75rem auto 0;
  max-width: 28ch;
  color: var(--muted);
  font-size: 0.75rem;
  line-height: 1.7;
  text-transform: lowercase;
}

.social-row,
.badge-row,
.stats {
  display: grid;
  gap: 0.7rem;
  width: 100%;
}

.social-row {
  display: inline-grid;
  width: auto;
  grid-auto-flow: column;
  gap: 0.7rem;
  margin: 0 auto;
}

.social-row__link,
.badge-row span,
.activity,
.stats article {
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.045);
}

.social-row__link {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 3.25rem;
  min-height: 3.25rem;
  gap: 0.35rem;
  padding: 0.65rem 0.75rem;
  border-radius: 18px;
  color: var(--text);
  transition: transform 180ms ease, background 180ms ease, border-color 180ms ease;
}

.social-row__link:hover {
  transform: translateY(-2px);
  border-color: rgb(var(--accent-rgb) / 0.28);
  background: rgb(var(--accent-rgb) / 0.1);
}

.social-row__icon {
  width: 1.7rem;
  height: 1.7rem;
  color: #f8f5f0;
  filter: drop-shadow(0 0 12px rgb(var(--accent-rgb) / 0.14));
}

.social-row__label {
  color: rgba(245, 243, 239, 0.9);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.72rem;
  letter-spacing: 0.08em;
  text-transform: lowercase;
}

.badge-row {
  grid-template-columns: repeat(3, 1fr);
}

.badge-row span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  border-radius: 999px;
  color: var(--muted);
  font-size: 0.64rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.activity {
  width: 100%;
  padding: 0.95rem 1rem;
  border-radius: 18px;
}

.activity__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  color: var(--muted);
  font-size: 0.62rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.activity__content {
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  gap: 0.85rem;
  align-items: center;
  margin-top: 0.7rem;
}

.activity__content--offline {
  grid-template-columns: 58px minmax(0, 1fr);
}

.activity__art {
  display: block;
  width: 58px;
  height: 58px;
  border-radius: 14px;
  object-fit: cover;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 10px 24px rgba(0, 0, 0, 0.18);
}

.activity__offline-orb {
  width: 58px;
  height: 58px;
  display: grid;
  place-items: center;
  border: 1px solid rgb(var(--warning-rgb) / 0.24);
  border-radius: 14px;
  background:
    linear-gradient(180deg, rgb(var(--warning-rgb) / 0.14), rgb(var(--warning-rgb) / 0.05)),
    rgba(255, 255, 255, 0.03);
  box-shadow:
    inset 0 1px 0 rgb(var(--warning-rgb) / 0.14),
    0 10px 24px rgba(0, 0, 0, 0.18);
}

.activity__offline-orb svg {
  width: 26px;
  height: 26px;
  color: var(--warning);
  filter: drop-shadow(0 0 10px rgb(var(--warning-rgb) / 0.2));
}

.activity__copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.16rem;
  min-width: 0;
  min-height: 58px;
}

.activity strong {
  display: block;
  margin: 0;
  color: var(--text);
  font-size: 0.96rem;
  line-height: 1.28;
  font-weight: 600;
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.activity__copy--offline .activity__title--offline {
  color: var(--warning-soft);
  line-height: 1.2;
}

.activity__artist {
  display: inline-block;
  margin-top: 0;
  color: var(--muted);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.68rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.activity__status {
  display: inline-flex;
  align-items: center;
  margin-top: 0;
}

.activity__status-label {
  color: var(--warning);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.stats {
  grid-template-columns: repeat(2, 1fr);
}

.stats article {
  display: grid;
  gap: 0.35rem;
  padding: 0.95rem 0.9rem;
  border-radius: 18px;
  text-align: center;
}

.stats strong {
  color: var(--text);
  font-size: 0.96rem;
  line-height: 1.4;
  word-break: break-word;
}

.stats span {
  color: var(--muted);
  font-size: 0.62rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 520px) {
  .scene {
    padding: 0.75rem;
  }

  .corner-player,
  .corner-views {
    padding: 0.75rem 0.8rem;
    border-radius: 18px;
  }

  .corner-player {
    top: 0.75rem;
    left: 0.75rem;
    min-width: 0;
  }

  .corner-views {
    right: 0.75rem;
    bottom: 0.75rem;
  }

  .profile-card {
    min-height: 650px;
    border-radius: 22px;
  }

  .profile-card--compact {
    min-height: 555px;
  }

  .profile-card--sparse {
    width: min(100%, 392px);
    min-height: 0;
  }

  .profile-card--minimal {
    width: min(100%, 356px);
    min-height: 0;
  }

  .profile-card__banner {
    height: 170px;
  }

  .profile-card__body {
    padding-inline: 0.95rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .profile,
  .social-row__link {
    transition: none;
  }
}

@media (pointer: fine) {
  .scene--custom-cursor,
  .scene--custom-cursor * {
    cursor: url('/assets/cursor/lionheart-pointer.cur') 1 1, auto;
  }

  .scene--custom-cursor a,
  .scene--custom-cursor button,
  .scene--custom-cursor input,
  .scene--custom-cursor textarea,
  .scene--custom-cursor select,
  .scene--custom-cursor [role='button'] {
    cursor: url('/assets/cursor/lionheart-pointer.cur') 1 1, pointer;
  }
}
