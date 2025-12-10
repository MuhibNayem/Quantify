import { writable } from 'svelte/store';
import api from '$lib/api';

export interface SystemSettings {
    currency_symbol: string;
    timezone: string;
    locale: string;
    business_name: string;
    return_window_days: number;
    tax_rate_percentage: number;
    loyalty_points_earning_rate: number;
    loyalty_points_redemption_rate: number;
    loyalty_tier_silver: number;
    loyalty_tier_gold: number;
    loyalty_tier_platinum: number;
}

const defaultSettings: SystemSettings = {
    currency_symbol: '$',
    timezone: 'UTC',
    locale: 'en-US',
    business_name: 'Quantify Business',
    return_window_days: 30,
    tax_rate_percentage: 0,
    loyalty_points_earning_rate: 1,
    loyalty_points_redemption_rate: 0.01,
    loyalty_tier_silver: 500,
    loyalty_tier_gold: 2500,
    loyalty_tier_platinum: 10000
};

function createSettingsStore() {
    const { subscribe, set, update } = writable<SystemSettings>(defaultSettings);

    return {
        subscribe,
        set,
        update,
        load: async () => {
            try {
                const response = await api.get('/settings/configurations');
                const data = response.data;

                const newSettings: SystemSettings = {
                    currency_symbol: data.currency_symbol || defaultSettings.currency_symbol,
                    timezone: data.timezone || defaultSettings.timezone,
                    locale: 'en-US',
                    business_name: data.business_name || defaultSettings.business_name,
                    return_window_days: parseFloat(data.return_window_days) || defaultSettings.return_window_days,
                    tax_rate_percentage: parseFloat(data.tax_rate_percentage) || defaultSettings.tax_rate_percentage,
                    loyalty_points_earning_rate: parseFloat(data.loyalty_points_earning_rate) || defaultSettings.loyalty_points_earning_rate,
                    loyalty_points_redemption_rate: parseFloat(data.loyalty_points_redemption_rate) || defaultSettings.loyalty_points_redemption_rate,
                    loyalty_tier_silver: parseFloat(data.loyalty_tier_silver) || defaultSettings.loyalty_tier_silver,
                    loyalty_tier_gold: parseFloat(data.loyalty_tier_gold) || defaultSettings.loyalty_tier_gold,
                    loyalty_tier_platinum: parseFloat(data.loyalty_tier_platinum) || defaultSettings.loyalty_tier_platinum
                };

                set(newSettings);

                // Update global config for utils
                updateGlobalConfig(newSettings);

            } catch (error) {
                console.error('Failed to load settings:', error);
            }
        }
    };
}

export const settings = createSettingsStore();

// -- Integration with Utils --
// We expose a way for utils.ts to read the current settings without subscribing
// This acts as a synchronous cache
export let globalConfig = { ...defaultSettings };

function updateGlobalConfig(newSettings: SystemSettings) {
    globalConfig = { ...newSettings };
}
