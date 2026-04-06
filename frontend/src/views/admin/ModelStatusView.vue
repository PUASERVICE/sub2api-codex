<template>
  <AppLayout>
    <div class="model-status-layout">
      <div class="card flex-1 min-h-0 overflow-hidden">
        <div class="model-status-shell">
          <a
            href="/status"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm model-status-open-fab"
          >
            {{ t('modelStatus.openInNewTab') }}
          </a>
          <iframe
            src="/status"
            class="model-status-frame"
            allowfullscreen
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
</script>

<style scoped>
.model-status-layout {
  @apply flex flex-col;
  height: calc(100vh - 64px - 4rem);
}

.model-status-shell {
  @apply relative;
  @apply h-full w-full overflow-hidden rounded-2xl;
  @apply bg-gradient-to-b from-gray-50 to-white dark:from-dark-900 dark:to-dark-950;
  @apply p-0;
}

.model-status-open-fab {
  @apply absolute right-3 top-3 z-10;
  @apply shadow-sm backdrop-blur supports-[backdrop-filter]:bg-white/80 dark:supports-[backdrop-filter]:bg-dark-800/80;
}

.model-status-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
</style>

