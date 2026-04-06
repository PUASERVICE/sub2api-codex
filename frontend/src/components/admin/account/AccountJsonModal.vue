<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.viewJsonTitle')"
    width="wide"
    @close="handleClose"
  >
    <div class="space-y-4">
      <div
        v-if="account"
        class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-gradient-to-r from-slate-50 to-gray-100 p-4 dark:border-dark-500 dark:from-dark-700 dark:to-dark-600"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-sky-500 to-cyan-600">
              <Icon name="document" size="md" class="text-white" />
            </div>
            <div>
              <div class="font-semibold text-gray-900 dark:text-gray-100">{{ account.name }}</div>
              <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                <span class="rounded bg-white/70 px-2 py-0.5 font-medium uppercase dark:bg-dark-500/80">
                  #{{ account.id }}
                </span>
                <span class="rounded bg-white/70 px-2 py-0.5 font-medium uppercase dark:bg-dark-500/80">
                  {{ account.platform }}
                </span>
                <span class="rounded bg-white/70 px-2 py-0.5 font-medium uppercase dark:bg-dark-500/80">
                  {{ account.type }}
                </span>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              @click="copyAccessToken"
              class="inline-flex items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm font-medium text-sky-700 transition-colors hover:bg-sky-100 disabled:cursor-not-allowed disabled:opacity-50 dark:border-sky-700/60 dark:bg-sky-900/20 dark:text-sky-200 dark:hover:bg-sky-900/30"
              :disabled="loading || !accessToken"
            >
              <Icon name="key" size="sm" />
              {{ t('admin.accounts.copyAccessToken') }}
            </button>
            <button
              @click="copyJson"
              class="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600"
              :disabled="loading || !prettyJson"
            >
              <Icon name="copy" size="sm" />
              {{ t('common.copy') }}
            </button>
          </div>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.accounts.viewJsonHint') }}
        </p>
      </div>

      <div
        v-if="loading"
        class="flex min-h-[320px] items-center justify-center rounded-xl border border-gray-200 bg-gray-50 text-sm text-gray-500 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-300"
      >
        {{ t('admin.accounts.viewJsonLoading') }}
      </div>

      <div
        v-else-if="errorMessage"
        class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300"
      >
        {{ errorMessage }}
      </div>

      <div
        v-else
        class="overflow-hidden rounded-xl border border-gray-200 bg-slate-950 dark:border-dark-500"
      >
        <pre class="max-h-[60vh] overflow-auto p-4 text-xs leading-6 text-slate-100"><code>{{ prettyJson }}</code></pre>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleClose"
          class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="loadAccountJson"
          class="inline-flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-600 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="loading || !account"
        >
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          {{ t('common.refresh') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { Icon } from '@/components/icons'
import { adminAPI } from '@/api/admin'
import { useClipboard } from '@/composables/useClipboard'
import type { Account } from '@/types'

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const fullAccount = ref<Account | null>(null)
const errorMessage = ref('')

const prettyJson = computed(() => {
  if (!fullAccount.value) return ''
  return JSON.stringify(fullAccount.value, null, 2)
})

const accessToken = computed(() => {
  const value = fullAccount.value?.credentials?.access_token
  return typeof value === 'string' ? value : ''
})

const loadAccountJson = async () => {
  if (!props.account?.id) return
  loading.value = true
  errorMessage.value = ''
  try {
    fullAccount.value = await adminAPI.accounts.getById(props.account.id)
  } catch (error: any) {
    console.error('Failed to load account JSON:', error)
    errorMessage.value = error?.message || t('admin.accounts.viewJsonLoadFailed')
  } finally {
    loading.value = false
  }
}

const copyJson = async () => {
  if (!prettyJson.value) return
  await copyToClipboard(prettyJson.value, t('admin.accounts.jsonCopied'))
}

const copyAccessToken = async () => {
  if (!accessToken.value) return
  await copyToClipboard(accessToken.value, t('admin.accounts.accessTokenCopied'))
}

const handleClose = () => {
  emit('close')
}

watch(
  () => [props.show, props.account?.id] as const,
  ([show, accountID], previousValue) => {
    const previousAccountID = previousValue?.[1]
    if (!show || !accountID) return
    if (fullAccount.value?.id === accountID && previousAccountID === accountID) return
    loadAccountJson()
  },
  { immediate: true }
)
</script>
