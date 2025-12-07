/**
 * Liquid Glass Design System
 * Single source of truth for Apple-inspired liquid glass aesthetics
 * 
 * Philosophy:
 * - Pure translucent glass with minimal tint
 * - High blur for depth and layering
 * - Adaptive text colors for readability
 * - Organic, flowing interactions
 */

// Core Glass Properties
export const liquidGlass = {
    // Background: Balanced translucency - visible yet glass-like
    background: {
        light: 'bg-white/[0.50]', // Light glass (50%)
        medium: 'bg-white/[0.65]', // Standard glass (65%)
        heavy: 'bg-white/[0.80]', // Solid glass (80%)
    },

    // Backdrop Blur: Heavy blur for authentic glass depth
    blur: {
        light: 'backdrop-blur-2xl',
        medium: 'backdrop-blur-3xl',
        heavy: 'backdrop-blur-[40px]',
    },

    // Borders: Clean, curved 3D edges (like mobile screen curves)
    border: {
        light: 'border-2 border-white/70',
        medium: 'border-2 border-white/90',
        heavy: 'border-2 border-white',
    },

    // 3D Depth Effect: Layered shadows for curved glass appearance
    innerGlow: {
        light: 'shadow-[inset_0_1px_0_0_rgba(255,255,255,0.5)]',
        medium: 'shadow-[inset_0_1px_0_0_rgba(255,255,255,0.7)]',
        heavy: 'shadow-[inset_0_2px_0_0_rgba(255,255,255,0.9)]',
    },

    // Shadows: Optical depth with subtle RGB values
    shadow: {
        light: 'shadow-[0_8px_32px_0_rgba(31,38,135,0.05)]',
        medium: 'shadow-[0_8px_32px_0_rgba(31,38,135,0.08)]',
        heavy: 'shadow-[0_8px_32px_0_rgba(31,38,135,0.12)]',
    },

    // Border Radius: Organic, flowing curves
    radius: {
        medium: 'rounded-3xl',
        large: 'rounded-[2rem]',
    },

    // Saturation: Enhanced color vibrancy
    saturate: 'backdrop-saturate-150',

    // Hover States: Realistic light reflection and bending
    hover: {
        shadow: 'hover:shadow-[0_8px_32px_0_rgba(31,38,135,0.18)]',
        border: 'hover:border-white',
        // Light reflection simulation
        reflection: 'hover:before:opacity-100 before:absolute before:inset-0 before:rounded-3xl before:bg-gradient-to-br before:from-white/40 before:via-transparent before:to-transparent before:opacity-0 before:transition-opacity before:duration-500',
        // Light bend on curves (shimmer effect)
        shimmer: 'hover:after:translate-x-full after:absolute after:inset-0 after:rounded-3xl after:-translate-x-full after:bg-gradient-to-r after:from-transparent after:via-white/20 after:to-transparent after:transition-transform after:duration-700 after:ease-out',
    },

    // Transitions: Smooth, liquid-like
    transition: 'transition-all duration-500 ease-out',
} as const;

// Text Colors: Adaptive based on background
export const adaptiveText = {
    // For use on glass surfaces
    onGlass: {
        primary: 'text-slate-900',
        secondary: 'text-slate-600',
        tertiary: 'text-slate-500',
        muted: 'text-slate-400',
    },

    // For headings
    heading: 'text-slate-900 font-bold tracking-tight',

    // For labels and metadata
    label: 'text-slate-500 text-sm font-medium',
} as const;

// Composite Glass Card Classes
export const glassCard = {
    base: [
        'relative overflow-hidden',
        liquidGlass.radius.medium,
        liquidGlass.border.medium,
        liquidGlass.background.medium,
        liquidGlass.blur.heavy,
        liquidGlass.saturate,
        liquidGlass.shadow.medium,
        liquidGlass.innerGlow.medium, // 3D curved border effect
        liquidGlass.transition,
        liquidGlass.hover.shadow,
        liquidGlass.hover.border,
        liquidGlass.hover.reflection, // Light reflection on hover
        liquidGlass.hover.shimmer, // Light bending shimmer
        'p-8',
    ].join(' '),

    modal: [
        'relative overflow-hidden',
        liquidGlass.radius.medium,
        liquidGlass.border.heavy,
        liquidGlass.background.heavy,
        liquidGlass.blur.heavy,
        liquidGlass.saturate,
        liquidGlass.shadow.heavy,
        liquidGlass.innerGlow.heavy, // 3D curved border effect
        'p-8',
    ].join(' '),
} as const;

// Background Gradients: Organic mesh patterns
export const meshGradient = {
    // Light, colorful ambient background
    ambient: `
		<div class="pointer-events-none absolute inset-0 overflow-hidden opacity-50">
			<div class="absolute left-[10%] top-[5%] h-[600px] w-[600px] rounded-full bg-gradient-to-br from-blue-200 via-cyan-100 to-transparent blur-[120px]"></div>
			<div class="absolute right-[5%] top-[30%] h-[500px] w-[500px] rounded-full bg-gradient-to-tr from-purple-200 via-pink-100 to-transparent blur-[100px]"></div>
			<div class="absolute bottom-[10%] left-[30%] h-[400px] w-[400px] rounded-full bg-gradient-to-tl from-indigo-200 via-violet-100 to-transparent blur-[90px]"></div>
		</div>
	`,
} as const;

// Helper function to combine glass classes
export function cn(...classes: (string | undefined | null | false)[]): string {
    return classes.filter(Boolean).join(' ');
}
