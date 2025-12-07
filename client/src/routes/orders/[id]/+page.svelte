<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import GlassCard from "$lib/components/ui/GlassCard.svelte";
  import { formatDate, formatCurrency } from "$lib/utils";
  import { fade, fly } from "svelte/transition";

  let orderId = $page.params.id;
  let order = $state(null);
  let loading = $state(true);
  let showReturnModal = $state(false);
  let returnReason = $state("");
  let selectedItems = $state({}); // { productId: quantity }

  // Mock Data for Detail
  const mockOrder = {
    id: 1,
    orderNumber: "ORD-1733600000-1",
    date: "2025-12-07T12:00:00Z",
    total: 150.00,
    status: "COMPLETED",
    items: [
      { id: 101, productId: 1, name: "Wireless Headphones", price: 100.00, quantity: 1, image: "https://placehold.co/100" },
      { id: 102, productId: 2, name: "USB-C Cable", price: 25.00, quantity: 2, image: "https://placehold.co/100" }
    ]
  };

  onMount(async () => {
    // Simulate fetch
    setTimeout(() => {
      order = mockOrder;
      loading = false;
    }, 500);
  });

  function toggleItemSelection(productId, maxQty) {
    if (selectedItems[productId]) {
      const newItems = { ...selectedItems };
      delete newItems[productId];
      selectedItems = newItems;
    } else {
      selectedItems = { ...selectedItems, [productId]: 1 };
    }
  }

  function updateQuantity(productId, qty, max) {
    if (qty > 0 && qty <= max) {
      selectedItems = { ...selectedItems, [productId]: qty };
    }
  }

  async function submitReturn() {
    const itemsToReturn = Object.entries(selectedItems).map(([pid, qty]) => ({
      product_id: parseInt(pid),
      quantity: qty,
      condition: "GOOD" // Default for now
    }));

    const payload = {
      order_number: order.orderNumber,
      items: itemsToReturn,
      reason: returnReason
    };

    console.log("Submitting Return:", payload);
    // Call API here: await api.post('/returns/request', payload);
    
    alert("Return request submitted!");
    showReturnModal = false;
  }
</script>

<div class="container mx-auto p-6 space-y-8">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600">
      Order Details
    </h1>
    <a href="/orders" class="text-gray-400 hover:text-white transition-colors">← Back to Orders</a>
  </div>

  {#if loading}
    <div class="text-center text-gray-400">Loading details...</div>
  {:else if order}
    <div class="grid gap-6 lg:grid-cols-3">
      <!-- Order Info -->
      <div class="lg:col-span-2 space-y-6">
        <GlassCard>
          <div class="flex justify-between items-start mb-6">
            <div>
              <h2 class="text-xl font-semibold text-white">Order #{order.orderNumber}</h2>
              <p class="text-sm text-gray-400">{formatDate(order.date)}</p>
            </div>
            <span class="text-green-400 bg-green-400/10 px-3 py-1 rounded-full text-sm">
              {order.status}
            </span>
          </div>

          <div class="space-y-4">
            {#each order.items as item}
              <div class="flex items-center justify-between p-4 rounded-lg bg-white/5 hover:bg-white/10 transition-colors">
                <div class="flex items-center gap-4">
                  <img src={item.image} alt={item.name} class="w-16 h-16 rounded-md object-cover" />
                  <div>
                    <h3 class="font-medium text-white">{item.name}</h3>
                    <p class="text-sm text-gray-400">Qty: {item.quantity} x {formatCurrency(item.price)}</p>
                  </div>
                </div>
                <span class="font-semibold text-white">{formatCurrency(item.price * item.quantity)}</span>
              </div>
            {/each}
          </div>
        </GlassCard>
      </div>

      <!-- Summary & Actions -->
      <div class="space-y-6">
        <GlassCard>
          <h3 class="text-lg font-semibold text-white mb-4">Summary</h3>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between text-gray-400">
              <span>Subtotal</span>
              <span>{formatCurrency(order.total)}</span>
            </div>
            <div class="flex justify-between text-gray-400">
              <span>Shipping</span>
              <span>Free</span>
            </div>
            <div class="border-t border-white/10 my-2 pt-2 flex justify-between text-white font-bold text-lg">
              <span>Total</span>
              <span>{formatCurrency(order.total)}</span>
            </div>
          </div>

          {#if order.status === 'COMPLETED'}
            <button 
              onclick={() => showReturnModal = true}
              class="w-full mt-6 bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-500 hover:to-pink-500 text-white font-semibold py-3 rounded-lg shadow-lg transition-all transform hover:scale-[1.02]"
            >
              Request Return
            </button>
          {/if}
        </GlassCard>
      </div>
    </div>
  {/if}

  <!-- Return Modal -->
  {#if showReturnModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" transition:fade>
      <div class="w-full max-w-lg" transition:fly={{ y: 20 }}>
        <GlassCard class="relative">
          <button 
            onclick={() => showReturnModal = false}
            class="absolute top-4 right-4 text-gray-400 hover:text-white"
          >
            ✕
          </button>
          
          <h2 class="text-2xl font-bold text-white mb-6">Request Return</h2>
          
          <div class="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
            <p class="text-gray-300 text-sm">Select items to return:</p>
            {#each order.items as item}
              <div class="flex items-center justify-between p-3 rounded-lg border border-white/10 bg-black/20">
                <div class="flex items-center gap-3">
                  <input 
                    type="checkbox" 
                    checked={!!selectedItems[item.productId]}
                    onchange={() => toggleItemSelection(item.productId, item.quantity)}
                    class="rounded border-gray-600 bg-gray-700 text-purple-600 focus:ring-purple-500"
                  />
                  <div>
                    <p class="text-white text-sm font-medium">{item.name}</p>
                    <p class="text-xs text-gray-500">{formatCurrency(item.price)}</p>
                  </div>
                </div>
                
                {#if selectedItems[item.productId]}
                  <div class="flex items-center gap-2">
                    <button 
                      class="w-6 h-6 rounded bg-white/10 hover:bg-white/20 text-white flex items-center justify-center"
                      onclick={() => updateQuantity(item.productId, selectedItems[item.productId] - 1, item.quantity)}
                    >-</button>
                    <span class="text-white text-sm w-4 text-center">{selectedItems[item.productId]}</span>
                    <button 
                      class="w-6 h-6 rounded bg-white/10 hover:bg-white/20 text-white flex items-center justify-center"
                      onclick={() => updateQuantity(item.productId, selectedItems[item.productId] + 1, item.quantity)}
                    >+</button>
                  </div>
                {/if}
              </div>
            {/each}

            <div class="pt-4">
              <label class="block text-sm font-medium text-gray-300 mb-2">Reason for Return</label>
              <textarea 
                bind:value={returnReason}
                class="w-full rounded-lg bg-black/20 border border-white/10 text-white p-3 focus:ring-2 focus:ring-purple-500 outline-none"
                rows="3"
                placeholder="Please describe why you are returning these items..."
              ></textarea>
            </div>
          </div>

          <div class="flex gap-3 mt-8">
            <button 
              onclick={() => showReturnModal = false}
              class="flex-1 py-2 rounded-lg border border-white/10 text-gray-300 hover:bg-white/5 transition-colors"
            >
              Cancel
            </button>
            <button 
              onclick={submitReturn}
              disabled={Object.keys(selectedItems).length === 0 || !returnReason}
              class="flex-1 py-2 rounded-lg bg-purple-600 text-white font-medium hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Submit Request
            </button>
          </div>
        </GlassCard>
      </div>
    </div>
  {/if}
</div>
