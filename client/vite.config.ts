import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
import { visualizer } from 'rollup-plugin-visualizer';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			registerType: 'autoUpdate',
			manifest: {
				name: 'Quantify',
				short_name: 'Quantify',
				description: 'Inventory, POS, and business intelligence',
				theme_color: '#0ea5e9',
				background_color: '#0f172a',
				display: 'standalone',
				scope: '/',
				start_url: '/',
				icons: [
					{ src: '/icon-192.png', sizes: '192x192', type: 'image/png' },
					{ src: '/icon-512.png', sizes: '512x512', type: 'image/png' },
					{
						src: '/icon-512-maskable.png',
						sizes: '512x512',
						type: 'image/png',
						purpose: 'maskable'
					}
				]
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,svg,woff2}'],
				runtimeCaching: [
					{
						urlPattern: ({ url }) => url.pathname.startsWith('/api/v1/'),
						method: 'GET',
						handler: 'StaleWhileRevalidate',
						options: {
							cacheName: 'api-cache',
							expiration: { maxEntries: 200, maxAgeSeconds: 24 * 60 * 60 }
						}
					}
				]
			}
		}),
		visualizer({ open: false, filename: 'dist/stats.html' })
	]
});
