import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	build: {
		// Suppress noisy a11y warnings during dev/build for speed.
		// The proper linter check is `pnpm run check` separately.
		reportCompressedSize: false
	},
	server: {
		fs: {
			allow: ['..']
		}
	}
});
