# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project in the current directory
npx sv create

# create a new project in my-app
npx sv create my-app
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## Image optimization

Prefer modern formats (WebP/AVIF) for new static images — they are typically 60–90% smaller than PNG.

Convert a PNG to WebP with [`cwebp`](https://developers.google.com/speed/webp/docs/cwebp):

```sh
cwebp -q 80 static/example.png -o static/example.webp
```

Serve with a `<picture>` fallback so older browsers still work:

```svelte
<picture>
	<source srcset="/example.webp" type="image/webp" />
	<img src="/example.png" alt="..." />
</picture>
```

## Bundle analysis

To visualize the production bundle and find large dependencies:

```sh
pnpm run analyze
```

This builds the app and writes a treemap report to `dist/stats.html`.
