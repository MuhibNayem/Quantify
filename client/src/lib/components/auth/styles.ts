// client/src/lib/components/auth/styles.ts
// Enterprise auth styles — porcelain card, indigo primary,
// Space Grotesk headings, precise 1px borders.

export const authCard =
	'surface-2 rounded-2xl p-7 sm:p-8 ' +
	'shadow-[0_24px_60px_-24px_rgb(15_23_42_/_0.18),0_4px_12px_-6px_rgb(15_23_42_/_0.08)]';

export const authInput = 'glass-input h-11 w-full pl-10 pr-3 text-sm';

export const authInputPlain = 'glass-input h-11 w-full px-3 text-sm';

export const authSelect =
	'glass-input flex h-11 w-full items-center justify-between px-3 text-sm ' +
	'disabled:cursor-not-allowed disabled:opacity-50';

export const authMenu =
	'surface-3 absolute top-full right-0 left-0 z-50 mt-2 overflow-hidden rounded-xl border border-border p-1.5';

export const authMenuItem =
	'flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm transition-colors';

export const authLabel =
	'text-[0.6875rem] font-semibold uppercase tracking-[0.1em] text-muted-foreground';

export const authCta = 'btn-soothe h-11 w-full rounded-lg px-4 text-sm';

export const authSecondaryBtn =
	'h-11 rounded-lg border border-border bg-card px-4 text-sm font-medium text-foreground ' +
	'shadow-xs transition-all hover:bg-muted hover:shadow-sm';

export const authLink =
	'font-medium text-primary underline-offset-4 transition-colors hover:underline ' +
	'hover:text-primary/80';

export const authError = 'text-xs font-medium text-danger';

export const authIcon =
	'pointer-events-none absolute top-1/2 left-3.5 h-4 w-4 -translate-y-1/2 text-muted-foreground';

export const authOverline =
	'text-[0.6875rem] font-semibold tracking-[0.16em] text-primary uppercase';

export const authTitle =
	'mt-3 font-display text-[2rem] leading-[1.12] font-semibold tracking-[-0.025em] ' +
	'text-foreground';

export const authSub = 'mt-3 text-[0.95rem] leading-relaxed text-muted-foreground';

export const authMobileTile =
	'flex h-10 w-10 items-center justify-center rounded-xl shadow-sm ' +
	'[background:var(--gradient-primary)]';

export const authTrustLine =
	'mt-4 flex items-center justify-center gap-1.5 text-[0.6875rem] text-muted-foreground';
