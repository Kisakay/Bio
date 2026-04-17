# Check what this Bio repositories look likes in production

### https://kisakay.com

# Hosting Your Own Profile with GitHub Pages

This repo is designed to publish the frontend on GitHub Pages through the workflow [deploy-pages.yml](/home/kisakay/Documents/Code/GitHub/Kisakay/.github/workflows/deploy-pages.yml).

Important:

- GitHub Pages only hosts the frontend
- the Go API in [server](/home/kisakay/Documents/Code/GitHub/Kisakay/server) cannot run on GitHub Pages
- you do not need to host your own API to use this profile
- the public API at `https://profile.kisakay.com` is meant to be reused and is the recommended stable option for Last.fm and the view counter

## 1. Fork the repo

Start by forking this repository to your own GitHub account.

Then clone your fork if you want to edit the project locally:

```bash
git clone https://github.com/YOUR_USERNAME/Kisakay.git
cd Kisakay
npm install
```

## 2. Customize the profile

The main file to edit is [src/config/config.ts](/home/kisakay/Documents/Code/GitHub/Kisakay/src/config/config.ts).

You can change:

- `content`: username, bio, title, pronouns, location
- `socials`: your links
- `badges`: displayed badges
- `features`: enable or disable sections
- `theme`: colors

You can also replace the files inside [public/assets](/home/kisakay/Documents/Code/GitHub/Kisakay/public/assets):

- `banner.jpg`
- `pfp.jpg`
- `song.mp3`
- `background.mp4`

## 3. Pick the right URL type

The most important setting for GitHub Pages is `VITE_BASE_PATH`.

This repo already uses that variable in [vite.config.ts](/home/kisakay/Documents/Code/GitHub/Kisakay/vite.config.ts).

### Case A: custom domain or user page

If you publish on:

- `https://your-username.github.io`
- or a custom domain such as `https://profile.example.com`

then use:

```env
VITE_BASE_PATH=/
```

### Case B: project page

If you publish on:

- `https://your-username.github.io/repository-name/`

then use:

```env
VITE_BASE_PATH=/repository-name/
```

## 4. Use the public API or disable it

The frontend reads `VITE_API_BASE_URL` to know where to call the API.

### Recommended setup

In most cases, you do not need to host any API yourself.

You can simply use the public project API:

```env
VITE_API_BASE_URL=https://profile.kisakay.com
```

This API is stable and is enough to power:

- the Last.fm display
- the view counter

In other words, if you want to host your own profile on GitHub Pages, only the frontend needs to be deployed.

### If you want a fully GitHub Pages-only site

Since GitHub Pages cannot run the Go API, you should disable dynamic features in [src/config/config.ts](/home/kisakay/Documents/Code/GitHub/Kisakay/src/config/config.ts):

```ts
features: {
  viewCounterEnabled: false,
  // ...
},
api: {
  lastfmEnabled: false,
  // ...
}
```

In that mode, you do not need any backend at all.

### If you want to keep Last.fm and the view counter

The simplest option is to keep using the public API:

```env
VITE_API_BASE_URL=https://profile.kisakay.com
```

So you do not need to deploy the `server/` folder.

Self-hosting the API is still possible if you want your own infrastructure, but it is not required to run your profile.

## 5. Update the GitHub Pages workflow

Deployment is handled by [deploy-pages.yml](/home/kisakay/Documents/Code/GitHub/Kisakay/.github/workflows/deploy-pages.yml).

In your fork, the main thing you need to adjust is `VITE_BASE_PATH`.

You can keep `VITE_API_BASE_URL` as-is to reuse the public API:

```yml
env:
  VITE_API_BASE_URL: https://profile.kisakay.com
  VITE_BASE_PATH: /
```

Examples:

- custom domain: `VITE_BASE_PATH: /`
- user page `your-username.github.io`: `VITE_BASE_PATH: /`
- project page `your-username.github.io/my-profile/`: `VITE_BASE_PATH: /my-profile/`

Only change `VITE_API_BASE_URL` if you explicitly want to connect your own API.

## 6. Manage the custom domain

The file [public/CNAME](/home/kisakay/Documents/Code/GitHub/Kisakay/public/CNAME) is used for a custom domain.

### If you use your own domain

Replace its contents with your domain:

```txt
profile.example.com
```

Then configure that same domain in the repository GitHub Pages settings.

### If you do not use a custom domain

Delete `public/CNAME` from your fork, otherwise GitHub Pages will try to use the original repo domain.

## 7. Enable GitHub Pages

In your GitHub repository:

1. go to `Settings`
2. open `Pages`
3. choose `GitHub Actions` as the deployment source if it is not already selected

Then push to `main` to trigger the workflow.

## 8. Test the build locally

Before pushing, you can verify the frontend build:

```bash
npm install
npm run build
```

If you want to test with a project-page base path, for example:

```bash
VITE_BASE_PATH=/repository-name/ npm run build
```

## 9. Quick summary

To host your own profile with this repo:

1. fork the repo
2. edit [src/config/config.ts](/home/kisakay/Documents/Code/GitHub/Kisakay/src/config/config.ts) and `public/assets`
3. set `VITE_BASE_PATH` for your GitHub Pages URL
4. replace or remove [public/CNAME](/home/kisakay/Documents/Code/GitHub/Kisakay/public/CNAME)
5. keep `VITE_API_BASE_URL=https://profile.kisakay.com` if you want to keep dynamic features
6. push to `main`

The key point is: GitHub Pages hosts the frontend, and `profile.kisakay.com` is enough for the API side in normal use.
