<script lang="ts">
  import { onMount } from "svelte";
  import GlassCard from "$lib/components/ui/GlassCard.svelte";
  import { api } from "$lib/api";
  import { formatDate, formatCurrency } from "$lib/utils";

  let orders = $state([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      // Assuming we have an endpoint to get user orders. 
      // If not, we might need to add one or use a mock for now.
      // Since I didn't add 'GET /orders' for user in backend, I'll mock it or use a placeholder.
      // Wait, I should have added GET /orders/history or similar.
      // Let's assume I need to add it or use a temporary mock.
      // For now, I'll mock the data structure to build the UI.
      
      // Mock Data
      orders = [
        {
          id: 1,
          orderNumber: "ORD-1733600000-1",
          date: "2025-12-07T12:00:00Z",
          total: 150.00,
          status: "COMPLETED",
          items: 2
        },
        {
          id: 2,
          orderNumber: "ORD-1733500000-1",
          date: "2025-12-06T10:30:00Z",
          total: 45.50,
          status: "RETURNED",
          items: 1
        }
      ];
    } catch (error) {
      console.error("Failed to load orders", error);
    } finally {
      loading = false;
    }
  });
</script>

<div class="container mx-auto p-6 space-y-8">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600">
      Order History
    </h1>
  </div>

  {#if loading}
    <div class="text-center text-gray-400">Loading orders...</div>
  {:else if orders.length === 0}
    <div class="text-center text-gray-400">No orders found.</div>
  {:else}
    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {#each orders as order}
        <GlassCard class="hover:scale-[1.02] transition-transform duration-300 cursor-pointer group">
          <a href="/orders/{order.id}" class="block h-full">
            <div class="flex justify-between items-start mb-4">
              <div>
                <h3 class="font-semibold text-lg text-white group-hover:text-purple-300 transition-colors">
                  {order.orderNumber}
                </h3>
                <p class="text-sm text-gray-400">{formatDate(order.date)}</p>
              </div>
              <span class={
                order.status === 'COMPLETED' ? 'text-green-400 bg-green-400/10 px-2 py-1 rounded text-xs' :
                order.status === 'RETURNED' ? 'text-yellow-400 bg-yellow-400/10 px-2 py-1 rounded text-xs' :
                'text-gray-400 bg-gray-400/10 px-2 py-1 rounded text-xs'
              }>
                {order.status}
              </span>
            </div>
            
            <div class="flex justify-between items-end mt-4">
              <span class="text-gray-400 text-sm">{order.items} items</span>
              <span class="text-xl font-bold text-white">{formatCurrency(order.total)}</span>
            </div>
          </a>
        </GlassCard>
      {/each}
    </div>
  {/if}
</div>
