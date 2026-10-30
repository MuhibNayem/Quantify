/**
 * Tailwind CSS v4 is configured CSS-first in `src/app.css`
 * (`@theme inline` + `@custom-variant dark`).
 *
 * This file is kept only for editor/tooling compatibility with older
 * Tailwind extensions. It intentionally declares no plugins or colors so it
 * can never drift from the real token system.
 */
module.exports = {
	darkMode: 'class',
	content: ['./src/**/*.{html,js,svelte,ts}']
};
