<!-- client/src/routes/login/+page.svelte -->
<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "$lib/components/ui/card";
  import api from "$lib/api";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";
  import { auth } from "$lib/stores/auth";
  import { onMount } from "svelte";

  let username = $state('');
  let password = $state('');
  let loading = $state(false);

  onMount(() => {
    auth.subscribe(state => {
      if (state.isAuthenticated) {
        goto('/');
      }
    })();
  });

  async function handleLogin() {
    loading = true;
    try {
      const response = await api.post('/users/login', { username, password });
      const { accessToken, refreshToken, csrfToken, user } = response.data;
      auth.login(accessToken, refreshToken, csrfToken, user);
      toast.success('Login successful!');
      goto('/');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Login failed';
      toast.error('Login Failed', {
        description: errorMessage,
      });
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex bg-gradient-to-br from-slate-50 to-blue-50 dark:from-slate-900 dark:to-slate-800">
  <!-- Left side - Image section -->
  <div class="hidden lg:flex flex-1 relative">
    <div 
      class="absolute inset-0 bg-cover bg-center"
      style="background-image: url('https://images.unsplash.com/photo-1551434678-e076c223a692?ixlib=rb-4.0.3&auto=format&fit=crop&w=2070&q=80')"
    ></div>
    <div class="absolute inset-0 bg-blue-600/10 dark:bg-blue-900/20"></div>
    <div class="relative z-10 flex flex-col justify-between p-12 text-white">
      <div class="flex items-center space-x-2">
        <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd" />
          </svg>
        </div>
        <span class="text-xl font-semibold">InventoryPro</span>
      </div>
      <div class="max-w-md">
        <h2 class="text-3xl font-bold mb-4">Streamline Your Inventory Management</h2>
        <p class="text-blue-100/90 leading-relaxed">
          Access your dashboard to manage products, track inventory, and optimize your business operations with our intuitive platform.
        </p>
      </div>
    </div>
  </div>

  <!-- Right side - Login form -->
  <div class="flex-1 flex items-center justify-center p-8">
    <div class="w-full max-w-md space-y-8">
      <!-- Mobile logo -->
      <div class="lg:hidden flex justify-center mb-8">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd" />
            </svg>
          </div>
          <span class="text-2xl font-bold text-slate-800 dark:text-white">InventoryPro</span>
        </div>
      </div>

      <Card class="border-0 shadow-xl rounded-2xl bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
        <CardHeader class="text-center pb-2 pt-8">
          <CardTitle class="text-2xl font-bold text-slate-800 dark:text-white">
            Welcome Back
          </CardTitle>
          <CardDescription class="text-slate-600 dark:text-slate-300">
            Sign in to continue to your dashboard
          </CardDescription>
        </CardHeader>
        <CardContent class="p-8 pt-4">
          <form onsubmit={(event) => { event.preventDefault(); handleLogin(); }} class="space-y-6">
            <div class="space-y-4">
              <div class="space-y-2">
                <Label for="username" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Username
                </Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="Enter your username"
                  bind:value={username}
                  required
                  class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-blue-500 focus:ring-blue-500 transition-colors"
                />
              </div>
              <div class="space-y-2">
                <Label for="password" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Password
                </Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  bind:value={password}
                  required
                  class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-blue-500 focus:ring-blue-500 transition-colors"
                />
              </div>
            </div>
            
            <Button 
              type="submit" 
              class="w-full h-12 rounded-lg bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50"
              disabled={loading}
            >
              {#if loading}
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Signing in...
              {:else}
                Sign in to your account
              {/if}
            </Button>
          </form>

          <div class="mt-6 text-center text-sm text-slate-600 dark:text-slate-400">
            Don't have an account?{' '}
            <a 
              href="/register" 
              class="font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 transition-colors"
            >
              Create one here
            </a>
          </div>
        </CardContent>
      </Card>

      <p class="text-xs text-center text-slate-500 dark:text-slate-400 mt-6">
        Secure login powered by modern authentication
      </p>
    </div>
  </div>
</div>

<style>
  :global(body) {
    background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  }
  
  @media (prefers-color-scheme: dark) {
    :global(body) {
      background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
  }
</style>
