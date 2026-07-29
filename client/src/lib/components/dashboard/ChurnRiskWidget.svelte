<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { crmApi } from '$lib/api/resources';
	import type { ChurnRisk, UserSummary } from '$lib/types';
    import { debounce } from '$lib/utils';
	import { Loader2, Users, Search, AlertTriangle, CheckCircle2, XCircle } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
    import { t } from '$lib/i18n';

	let loading = false;
	let userId: number | null = null;
    let customerSearch = '';
    let searchResults: UserSummary[] = [];
    let selectedCustomer: UserSummary | null = null;
    let searching = false;

	let churnRisk: ChurnRisk | null = null;

    const searchCustomers = debounce(async () => {
        if (customerSearch.length < 2) return;
        searching = true;
        try {
            const res = await crmApi.listCustomers({ q: customerSearch, limit: 5 });
            searchResults = res.users;
        } catch (e) {
            console.error(e);
        } finally {
            searching = false;
        }
    }, 300);

    function selectCustomer(customer: UserSummary) {
        selectedCustomer = customer;
        userId = customer.ID;
        searchResults = [];
        customerSearch = '';
        churnRisk = null;
    }

	async function analyzeChurnRisk() {
		if (!userId) {
			toast.error($t('intelligence.toasts.select_customer'));
			return;
		}

		loading = true;
		churnRisk = null;
		try {
			const res = await crmApi.getChurnRisk(userId);
			churnRisk = res;
			toast.success($t('intelligence.toasts.analysis_complete'));
		} catch (error: any) {
			toast.error($t('intelligence.toasts.analysis_fail'), {
				description: error?.response?.data?.error || 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	}

    function getRiskColor(level: string) {
        switch (level?.toLowerCase()) {
            case 'high': return 'text-danger bg-danger/10 border-danger/25';
            case 'medium': return 'text-warning bg-warning/10 border-warning/25';
            case 'low': return 'text-success bg-success/10 border-success/25';
            default: return 'text-muted-foreground bg-muted border-border';
        }
    }
</script>

<Card class="h-full border-border bg-card shadow-sm">
	<CardHeader>
		<CardTitle class="flex items-center gap-2 text-foreground">
			<div class="rounded-md bg-muted p-2 text-primary">
				<Users class="h-5 w-5" />
			</div>
			{$t('intelligence.churn_risk.title')}
		</CardTitle>
		<CardDescription>{$t('intelligence.churn_risk.subtitle')}</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
        <!-- Customer Search -->
		<div class="space-y-2">
			<Label>{$t('intelligence.churn_risk.select_customer')}</Label>
            {#if selectedCustomer}
                <div class="flex items-center justify-between rounded-md border border-border bg-muted p-2">
                    <div class="flex flex-col">
                        <span class="font-medium text-foreground">{selectedCustomer.FirstName} {selectedCustomer.LastName}</span>
                        <span class="text-xs text-muted-foreground">{selectedCustomer.Email}</span>
                    </div>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 text-muted-foreground hover:text-foreground" onclick={() => { selectedCustomer = null; userId = null; churnRisk = null; }}>
                        X
                    </Button>
                </div>
            {:else}
                <div class="relative">
                    <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder={$t('intelligence.churn_risk.placeholder')}
                        class="pl-9"
                        bind:value={customerSearch}
                        oninput={searchCustomers}
                    />
                    {#if searchResults.length > 0}
                        <div class="absolute z-10 mt-1 w-full rounded-md border border-border bg-card shadow-lg">
                            {#each searchResults as customer}
                                <button
                                    class="w-full px-4 py-2 text-left text-sm transition-colors hover:bg-muted"
                                    onclick={() => selectCustomer(customer)}
                                >
                                    <div class="font-medium text-foreground">{customer.FirstName} {customer.LastName}</div>
                                    <div class="text-xs text-muted-foreground">{customer.Email}</div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
		</div>

		<div class="flex justify-end">
			<Button onclick={analyzeChurnRisk} disabled={loading || !userId}>
				{#if loading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					{$t('intelligence.churn_risk.analyzing_btn')}
				{:else}
					{$t('intelligence.churn_risk.analyze_btn')}
				{/if}
			</Button>
		</div>

		{#if churnRisk}
			<div class="mt-4 space-y-4 rounded-lg border border-border bg-card p-4 shadow-sm animate-in fade-in slide-in-from-bottom-4">
				<div class="flex items-center justify-between">
					<div>
						<div class="overline">{$t('intelligence.churn_risk.risk_level')}</div>
                        <div class={`mt-1 inline-flex items-center rounded-md border px-2.5 py-0.5 text-sm font-semibold ${getRiskColor(churnRisk.risk_level)}`}>
                            {churnRisk.risk_level}
                        </div>
					</div>
                    <div class="text-right">
                        <div class="overline">{$t('intelligence.churn_risk.risk_score')}</div>
                        <div class="kpi-num text-2xl text-foreground">{(churnRisk.churn_risk_score * 100).toFixed(0)}%</div>
                    </div>
				</div>

                {#if churnRisk.primary_factors && churnRisk.primary_factors.length > 0}
                    <div class="space-y-1">
                        <div class="text-sm font-medium text-foreground">{$t('intelligence.churn_risk.primary_factors')}</div>
                        <ul class="list-inside list-disc text-sm text-muted-foreground">
                            {#each churnRisk.primary_factors as factor}
                                <li>{factor}</li>
                            {/each}
                        </ul>
                    </div>
                {/if}

                <div class="rounded-md bg-muted p-3 text-sm text-foreground">
                    <div class="mb-1 font-medium flex items-center gap-2">
                        <CheckCircle2 class="h-3 w-3 text-success" /> {$t('intelligence.churn_risk.retention_strategy')}
                    </div>
                    {churnRisk.retention_strategy}
                </div>

                {#if churnRisk.suggested_discount > 0}
                    <div class="rounded-md border border-success/25 bg-success/10 p-3 text-sm text-success">
                        <div class="font-medium flex items-center gap-2">
                            <AlertTriangle class="h-3 w-3" /> {$t('intelligence.churn_risk.suggested_action')}
                        </div>
                        {$t('intelligence.churn_risk.discount_offer', { discount: churnRisk.suggested_discount })}
                    </div>
                {/if}
			</div>
		{/if}
	</CardContent>
</Card>
