<!-- client/src/routes/register/+page.svelte -->
<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "$lib/components/ui/card";
  import Select from "$lib/components/ui/select/select.svelte";
  import api from "$lib/api";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";

  let username = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let firstName = $state('');
  let lastName = $state('');
  let email = $state('');
  let phoneNumber = $state('');
  let address = $state('');
  let selectedRole = $state<'Admin' | 'Manager' | 'Staff'>('Staff'); // Default role
  let availableRoles: { value: string; label: string }[] = [
    { value: 'Admin', label: 'Administrator' },
    { value: 'Manager', label: 'Manager' },
    { value: 'Staff', label: 'Staff Member' },
  ];
  let loading = $state(false);
  let passwordStrength = $state(0);

  function calculatePasswordStrength(password: string) {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    return strength;
  }

  $effect(() => {
    passwordStrength = calculatePasswordStrength(password);
  });

  async function handleRegister() {
    loading = true;
    
    if (password !== confirmPassword) {
      toast.error('Registration Failed', {
        description: 'Passwords do not match.',
      });
      loading = false;
      return;
    }

    if (passwordStrength < 75) {
      toast.error('Registration Failed', {
        description: 'Please choose a stronger password.',
      });
      loading = false;
      return;
    }

    try {
      await api.post('/users/register', { 
        username, 
        password, 
        role: selectedRole,
        firstName: firstName || undefined,
        lastName: lastName || undefined,
        email: email || undefined,
        phoneNumber: phoneNumber || undefined,
        address: address || undefined
      });
      toast.success('Registration successful! Please log in.');
      goto('/login');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Registration failed';
      toast.error('Registration Failed', {
        description: errorMessage,
      });
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex bg-gradient-to-br from-slate-50 to-emerald-50 dark:from-slate-900 dark:to-slate-800">
  <!-- Left side - Image section -->
  <div class="hidden lg:flex flex-1 relative">
    <div 
      class="absolute inset-0 bg-cover bg-center"
      style="background-image: url('https://images.unsplash.com/photo-1556740758-90de374c12ad?ixlib=rb-4.0.3&auto=format&fit=crop&w=2070&q=80')"
    ></div>
    <div class="absolute inset-0 bg-emerald-600/10 dark:bg-emerald-900/20"></div>
    <div class="relative z-10 flex flex-col justify-between p-12 text-white">
      <div class="flex items-center space-x-2">
        <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
          </svg>
        </div>
        <span class="text-xl font-semibold">InventoryPro</span>
      </div>
      <div class="max-w-md">
        <h2 class="text-3xl font-bold mb-4">Join Our Platform</h2>
        <p class="text-emerald-100/90 leading-relaxed">
          Create your account to start managing inventory efficiently. Choose the right role for your responsibilities and get started in minutes.
        </p>
      </div>
      <div class="space-y-3">
        <div class="flex items-center space-x-3">
          <div class="w-6 h-6 bg-emerald-500 rounded-full flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </div>
          <span class="text-sm text-emerald-100/90">Secure account creation</span>
        </div>
        <div class="flex items-center space-x-3">
          <div class="w-6 h-6 bg-emerald-500 rounded-full flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </div>
          <span class="text-sm text-emerald-100/90">Role-based access control</span>
        </div>
        <div class="flex items-center space-x-3">
          <div class="w-6 h-6 bg-emerald-500 rounded-full flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </div>
          <span class="text-sm text-emerald-100/90">Instant access after approval</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Right side - Registration form -->
  <div class="flex-1 flex items-center justify-center p-8">
    <div class="w-full max-w-md space-y-8">
      <!-- Mobile logo -->
      <div class="lg:hidden flex justify-center mb-8">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-emerald-600 rounded-xl flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
            </svg>
          </div>
          <span class="text-2xl font-bold text-slate-800 dark:text-white">InventoryPro</span>
        </div>
      </div>

      <Card class="border-0 shadow-xl rounded-2xl bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
        <CardHeader class="text-center pb-2 pt-8">
          <CardTitle class="text-2xl font-bold text-slate-800 dark:text-white">
            Create Account
          </CardTitle>
          <CardDescription class="text-slate-600 dark:text-slate-300">
            Join us to manage inventory efficiently
          </CardDescription>
        </CardHeader>
        <CardContent class="p-8 pt-4">
          <form onsubmit={(event) => { event.preventDefault(); handleRegister(); }} class="space-y-6">
            <div class="space-y-4">
              <div class="space-y-2">
                <Label for="username" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Username
                </Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="Choose a username"
                  bind:value={username}
                  required
                  class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
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
                  class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                />
                {#if password}
                  <div class="space-y-2">
                    <div class="flex justify-between text-xs text-slate-600 dark:text-slate-400">
                      <span>Password strength</span>
                      <span>{passwordStrength}%</span>
                    </div>
                    <div class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                      <div 
                        class="h-2 rounded-full transition-all duration-300"
                        class:bg-red-500={passwordStrength < 50}
                        class:bg-yellow-500={passwordStrength >= 50 && passwordStrength < 75}
                        class:bg-green-500={passwordStrength >= 75}
                        style={`width: ${passwordStrength}%`}
                      ></div>
                    </div>
                  </div>
                {/if}
              </div>

              <div class="space-y-2">
                <Label for="confirmPassword" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Confirm Password
                </Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="••••••••"
                  bind:value={confirmPassword}
                  required
                  class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                />
                {#if confirmPassword && password !== confirmPassword}
                  <p class="text-red-500 text-xs flex items-center space-x-1">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                    </svg>
                    <span>Passwords do not match</span>
                  </p>
                {/if}
              </div>

              <div class="space-y-2">
                <Label for="role" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Role
                </Label>
                <select
                  class="h-12 w-full rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors text-sm appearance-none px-4"
                  bind:value={selectedRole}
                  disabled={loading}
                >
                  {#each availableRoles as role}
                    <option value={role.value}>{role.label}</option>
                  {/each}
                </select>
                
                <p class="text-xs text-slate-500 dark:text-slate-400">
                  {#if selectedRole === 'Admin'}
                    Full system access with all permissions
                  {:else if selectedRole === 'Manager'}
                    Manage teams and inventory with limited admin access
                  {:else}
                    Standard user access for daily operations
                  {/if}
                </p>
              </div>

              <!-- Personal Information Section -->
              <div class="pt-4 border-t border-slate-200 dark:border-slate-700">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">Personal Information (Optional)</h3>
                
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="space-y-2">
                    <Label for="firstName" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                      First Name
                    </Label>
                    <Input
                      id="firstName"
                      type="text"
                      placeholder="John"
                      bind:value={firstName}
                      class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                    />
                  </div>

                  <div class="space-y-2">
                    <Label for="lastName" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                      Last Name
                    </Label>
                    <Input
                      id="lastName"
                      type="text"
                      placeholder="Doe"
                      bind:value={lastName}
                      class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                    />
                  </div>
                </div>

                <div class="space-y-2 mt-4">
                  <Label for="email" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                    Email
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="john.doe@example.com"
                    bind:value={email}
                    class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                  />
                </div>

                <div class="space-y-2 mt-4">
                  <Label for="phoneNumber" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                    Phone Number
                  </Label>
                  <Input
                    id="phoneNumber"
                    type="tel"
                    placeholder="+1 (555) 123-4567"
                    bind:value={phoneNumber}
                    class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                  />
                </div>

                <div class="space-y-2 mt-4">
                  <Label for="address" class="text-sm font-medium text-slate-700 dark:text-slate-300">
                    Address
                  </Label>
                  <Input
                    id="address"
                    type="text"
                    placeholder="123 Main Street, City, State"
                    bind:value={address}
                    class="h-12 rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 focus:border-emerald-500 focus:ring-emerald-500 transition-colors"
                  />
                </div>
              </div>
            </div>
            
            <Button 
              type="submit" 
              class="w-full h-12 rounded-lg bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-700 hover:to-emerald-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50"
              disabled={loading ?? (password && confirmPassword && password !== confirmPassword)}
            >
              {#if loading}
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Creating account...
              {:else}
                Create account
              {/if}
            </Button>
          </form>

          <div class="mt-6 text-center text-sm text-slate-600 dark:text-slate-400">
            Already have an account?{' '}
            <a 
              href="/login" 
              class="font-semibold text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 dark:hover:text-emerald-300 transition-colors"
            >
              Sign in here
            </a>
          </div>
        </CardContent>
      </Card>

      <p class="text-xs text-center text-slate-500 dark:text-slate-400 mt-6">
        Your account will be ready immediately after registration
      </p>
    </div>
  </div>
</div>

<style>
  :global(body) {
    background: linear-gradient(135deg, #f8fafc 0%, #d1fae5 100%);
  }
  
  @media (prefers-color-scheme: dark) {
    :global(body) {
      background: linear-gradient(135deg, #0f172a 0%, #064e3b 100%);
    }
  }
</style>
