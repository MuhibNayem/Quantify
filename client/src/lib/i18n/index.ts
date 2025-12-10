import { derived } from 'svelte/store';
import { settings } from '$lib/stores/settings';
import enUS from './locales/en-US';
import bnBD from './locales/bn-BD';

const dictionaries: Record<string, any> = {
    'en-US': enUS,
    'bn-BD': bnBD
};

// Start with English (fallback)
let currentDictionary = enUS;

// Helper to access nested keys like 'settings.tabs.general'
function gl(keys: string, dict: any): string | undefined {
    return keys.split('.').reduce((acc, key) => acc && acc[key], dict);
}

export const t = derived(settings, ($settings) => {
    const locale = $settings.locale || 'en-US';
    const dict = dictionaries[locale] || dictionaries['en-US'];

    return (key: string, vars: Record<string, any> = {}) => {
        let text = gl(key, dict) || gl(key, dictionaries['en-US']) || key;

        // Simple variable substitution {var}
        Object.keys(vars).forEach((k) => {
            const regex = new RegExp(`{${k}}`, 'g');
            text = text.replace(regex, vars[k]);
        });

        return text;
    };
});
