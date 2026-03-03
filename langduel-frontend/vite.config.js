import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/ws': {
				target: 'ws://localhost:8080',
				ws: true,
				changeOrigin: true
			},
			'/auth': {
				target: 'http://localhost:8080',
				changeOrigin: true
			},
			'/me': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
