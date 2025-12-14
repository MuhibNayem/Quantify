import { derived } from 'svelte/store';
import { settings } from '$lib/stores/settings';
import enUS from './locales/en-US';
import bnBD from './locales/bn-BD';
import hiIN from './locales/hi-IN';
import jaJP from './locales/ja-JP';
import itIT from './locales/it-IT';
import esES from './locales/es-ES';
import arSA from './locales/ar-SA';

const dictionaries: Record<string, any> = {
    'en-US': enUS,
    'bn-BD': bnBD,
    'hi-IN': hiIN,
    'ja-JP': jaJP,
    'it-IT': itIT,
    'es-ES': esES,
    'ar-SA': arSA
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
