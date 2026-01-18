<script setup lang="ts">
import AppLayout from "@/components/layout/AppLayout.vue";
import Login from "./views/Login.vue";
import IdentitySetup from "./views/IdentitySetup.vue";
import {useAuthStore} from "./stores/auth";

import VaultSetup from "@/components/VaultSetup.vue";
import VaultUnlock from "@/components/VaultUnlock.vue";
import { useVaultStore } from "@/stores/vault";
import { ref, watch, computed } from "vue";
import { IdentityService } from "@/services/identity.service";
import { Sonner } from "@/components/ui/sonner";

const auth = useAuthStore();
const vault = useVaultStore();

const isInitializing = ref(true);
const isRestoringSession = ref(false);
const sessionRestoreFailed = ref(false);

// Check if vault has identity (keypair)
const hasKeypair = computed(() => !!vault.publicKey && !!vault.privateKey);

// Check if master identity key is loaded in memory
const hasIdentityInMemory = computed(() => !!IdentityService.getMasterKey());

// Check if we have a persisted user (returning user scenario)
const hasPersistedUser = computed(() => !!auth.user?.id);

// Determine if we need to show login
// Show login if: no persisted user, OR session restore failed
const needsLogin = computed(() => {
    if (sessionRestoreFailed.value) return true;
    if (!hasPersistedUser.value) return true;
    return false;
});

/**
 * SESSION RESTORE FLOW:
 * 1. If user is persisted (returning user) → Initialize vault, show VaultUnlock
 * 2. After vault unlock → Try to restore session with refresh token
 * 3. If restore succeeds → Continue to app
 * 4. If restore fails → Show Login (but keep vault intact)
 * 5. If no persisted user → Show Login
 */

async function initVaultForUser() {
    if (!auth.user?.id) return;

    // Use user-specific vault
    vault.setUserId(auth.user.id);
    await vault.checkStatus();
}

// Watch for vault unlock to restore session or sync data
watch(() => vault.status, async (newStatus) => {
    if (newStatus === 'unlocked') {
        // If we have tokens already, just sync
        if (auth.isAuthenticated) {
            await auth.persistPendingRefreshToken();
            if (vault.publicKey) {
                auth.updatePublicKey(vault.publicKey);
            }
        } else if (hasPersistedUser.value && !sessionRestoreFailed.value) {
            // Try to restore session using refresh token from vault
            isRestoringSession.value = true;
            const success = await auth.tryRestoreSession();
            isRestoringSession.value = false;

            if (success) {
                // Session restored! Sync public key
                if (vault.publicKey) {
                    auth.updatePublicKey(vault.publicKey);
                }
            } else {
                // Refresh token invalid/expired - need to re-login
                sessionRestoreFailed.value = true;
            }
        }
    }
});

function onIdentityCompleted() {
    // Identity setup completed - keys are now in vault
    // Sync public key to backend
    if (vault.publicKey) {
        auth.updatePublicKey(vault.publicKey);
    }
}

// Called after successful login to reset the failed state
function onLoginSuccess() {
    sessionRestoreFailed.value = false;
}

// Initialize
async function init() {
    isInitializing.value = true;

    if (hasPersistedUser.value) {
        // Returning user - initialize their vault
        await initVaultForUser();
    }

    isInitializing.value = false;
}

init();
</script>

<template>
  <Sonner />
  <div class="font-sans antialiased text-foreground bg-background min-h-screen">
      <!-- 0. Initializing -->
      <div v-if="isInitializing" class="flex items-center justify-center h-screen">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>

      <!-- 1. Restoring session -->
      <div v-else-if="isRestoringSession" class="flex flex-col items-center justify-center h-screen gap-4">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <p class="text-muted-foreground">Restoring session...</p>
      </div>

      <!-- 2. Needs Login (no persisted user OR session restore failed) -->
      <div v-else-if="needsLogin">
          <Login @success="onLoginSuccess" />
      </div>

      <!-- 3. Returning user - Vault flow before authentication -->
      <div v-else-if="hasPersistedUser && !auth.isAuthenticated">
          <!-- 3a. Vault loading -->
          <div v-if="vault.status === 'loading'" class="flex items-center justify-center h-screen">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <!-- 3b. Vault locked - need to unlock to get refresh token -->
          <div v-else-if="vault.status === 'locked'">
              <VaultUnlock :user-name="auth.user?.name" />
          </div>

          <!-- 3c. Vault uninitialized - shouldn't happen for returning user, show login -->
          <div v-else-if="vault.status === 'uninitialized'">
              <Login @success="onLoginSuccess" />
          </div>
      </div>

      <!-- 4. Authenticated → Vault & Identity Flow -->
      <div v-else-if="auth.isAuthenticated">
          <!-- 4a. Vault loading -->
          <div v-if="vault.status === 'loading'" class="flex items-center justify-center h-screen">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <!-- 4b. Vault Setup (new user on this device) -->
          <div v-else-if="vault.status === 'uninitialized'">
              <VaultSetup />
          </div>

          <!-- 4c. Vault Unlock (returning user) -->
          <div v-else-if="vault.status === 'locked'">
              <VaultUnlock :user-name="auth.user?.name" />
          </div>

          <!-- 4d. Vault Unlocked -->
          <div v-else-if="vault.status === 'unlocked'">
              <!-- No keypair yet - Identity Setup -->
              <div v-if="!hasKeypair">
                  <IdentitySetup @completed="onIdentityCompleted" />
              </div>

              <!-- Has keypair but no master key in memory - Identity Setup (recovery) -->
              <div v-else-if="!hasIdentityInMemory">
                  <IdentitySetup @completed="onIdentityCompleted" />
              </div>

              <!-- Everything ready - App -->
              <AppLayout v-else>
                  <router-view></router-view>
              </AppLayout>
          </div>
      </div>
  </div>
</template>
