<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
        <div class="card p-5">
          <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.total') }}
          </div>
          <div class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ overview.total_targets }}
          </div>
          <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.enabled', { count: overview.enabled_targets }) }}
          </div>
        </div>

        <div class="card p-5">
          <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.healthy') }}
          </div>
          <div class="mt-3 text-3xl font-semibold text-emerald-600 dark:text-emerald-400">
            {{ overview.healthy_targets }}
          </div>
          <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ healthyRatio }}
          </div>
        </div>

        <div class="card p-5">
          <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.failed') }}
          </div>
          <div class="mt-3 text-3xl font-semibold text-red-600 dark:text-red-400">
            {{ overview.failed_targets }}
          </div>
          <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.needAttention') }}
          </div>
        </div>

        <div class="card p-5">
          <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.unknown') }}
          </div>
          <div class="mt-3 text-3xl font-semibold text-amber-600 dark:text-amber-400">
            {{ overview.unknown_targets }}
          </div>
          <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.waiting') }}
          </div>
        </div>

        <div class="card p-5">
          <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
            {{ t('admin.modelStatus.summary.latency') }}
          </div>
          <div class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ averageLatencyLabel }}
          </div>
          <div class="mt-1 truncate text-sm text-gray-500 dark:text-gray-400">
            {{ overview.last_checked_target || t('common.unknown') }}
          </div>
        </div>
      </section>

      <section class="card p-5">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ isEditing ? t('admin.modelStatus.editTarget') : t('admin.modelStatus.createTarget') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.modelStatus.formHint') }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button class="btn btn-secondary" :disabled="loading" @click="refreshAll">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" class="mr-1.5" />
              {{ t('common.refresh') }}
            </button>
            <button
              v-if="isEditing"
              class="btn btn-secondary"
              type="button"
              @click="resetForm"
            >
              {{ t('common.cancel') }}
            </button>
            <button
              class="btn btn-primary"
              :disabled="submitting || !form.account_id || !form.model_id.trim()"
              @click="submitForm"
            >
              <Icon name="plus" size="sm" class="mr-1.5" />
              {{ isEditing ? t('common.save') : t('common.create') }}
            </button>
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          <div class="xl:col-span-2">
            <label class="input-label">{{ t('common.name') }}</label>
            <input v-model="form.name" type="text" class="input" :placeholder="t('admin.modelStatus.namePlaceholder')" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelStatus.account') }}</label>
            <select v-model.number="form.account_id" class="input">
              <option :value="0">{{ t('admin.modelStatus.selectAccount') }}</option>
              <option v-for="account in accountOptions" :key="account.id" :value="account.id">
                {{ account.name }} · {{ account.platform }}
              </option>
            </select>
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelStatus.model') }}</label>
            <input v-model="form.model_id" type="text" class="input" :placeholder="t('admin.modelStatus.modelPlaceholder')" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelStatus.interval') }}</label>
            <input v-model.number="form.check_interval_seconds" type="number" min="60" step="60" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelStatus.timeout') }}</label>
            <input v-model.number="form.timeout_seconds" type="number" min="5" max="300" step="5" class="input" />
          </div>
          <div class="flex items-center gap-2 pt-7">
            <input id="target-enabled" v-model="form.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
            <label for="target-enabled" class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('admin.modelStatus.enabled') }}
            </label>
          </div>
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.modelStatus.targets') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.modelStatus.targetsHint') }}
            </p>
          </div>
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input v-model="includeDisabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
            {{ t('admin.modelStatus.showDisabled') }}
          </label>
        </div>

        <div v-if="loading" class="p-8 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.loading') }}
        </div>

        <EmptyState
          v-else-if="targets.length === 0"
          :title="t('admin.modelStatus.emptyTitle')"
          :description="t('admin.modelStatus.emptyDescription')"
        />

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50/80 dark:bg-dark-800/70">
              <tr>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.name') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('admin.modelStatus.account') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('admin.modelStatus.model') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.status') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('admin.modelStatus.latency') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('admin.modelStatus.lastChecked') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('admin.modelStatus.nextCheck') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr
                v-for="target in targets"
                :key="target.id"
                class="align-top transition-colors hover:bg-gray-50/70 dark:hover:bg-dark-800/40"
              >
                <td class="px-4 py-4">
                  <div class="font-medium text-gray-900 dark:text-white">{{ target.name }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.modelStatus.failures', { count: target.consecutive_failures }) }}
                  </div>
                </td>
                <td class="px-4 py-4">
                  <div class="text-gray-900 dark:text-white">{{ target.account_name || `#${target.account_id}` }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ target.account_platform || '-' }} · {{ target.account_status || '-' }}
                  </div>
                </td>
                <td class="px-4 py-4">
                  <code class="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-800 dark:bg-dark-700 dark:text-gray-200">
                    {{ target.model_id }}
                  </code>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.modelStatus.intervalValue', { seconds: target.check_interval_seconds }) }}
                  </div>
                </td>
                <td class="px-4 py-4">
                  <span :class="statusBadgeClass(target.latest_status)">
                    {{ statusLabel(target.latest_status) }}
                  </span>
                  <div v-if="!target.enabled" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('common.disabled') }}
                  </div>
                  <div v-else-if="target.latest_error_message" class="mt-2 max-w-xs whitespace-pre-wrap break-words text-xs text-red-600 dark:text-red-400">
                    {{ target.latest_error_message }}
                  </div>
                </td>
                <td class="px-4 py-4 text-gray-700 dark:text-gray-300">
                  {{ target.latest_latency_ms != null ? `${target.latest_latency_ms} ms` : '-' }}
                </td>
                <td class="px-4 py-4 text-gray-600 dark:text-gray-300">
                  <div>{{ target.last_checked_at ? formatDateTime(target.last_checked_at) : '-' }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ target.last_success_at ? t('admin.modelStatus.lastSuccessAt', { time: formatDateTime(target.last_success_at) }) : t('admin.modelStatus.noSuccessYet') }}
                  </div>
                </td>
                <td class="px-4 py-4 text-gray-600 dark:text-gray-300">
                  {{ target.next_check_at ? formatDateTime(target.next_check_at) : '-' }}
                </td>
                <td class="px-4 py-4">
                  <div class="flex flex-wrap gap-2">
                    <button class="btn btn-secondary btn-sm" @click="startEdit(target)">
                      {{ t('common.edit') }}
                    </button>
                    <button class="btn btn-secondary btn-sm" :disabled="runningId === target.id" @click="handleRun(target.id)">
                      {{ runningId === target.id ? t('common.loading') : t('admin.modelStatus.runNow') }}
                    </button>
                    <button class="btn btn-secondary btn-sm" @click="toggleEnabled(target)">
                      {{ target.enabled ? t('common.disable') : t('common.enable') }}
                    </button>
                    <button class="btn btn-secondary btn-sm" @click="selectHistoryTarget(target)">
                      {{ t('admin.modelStatus.history') }}
                    </button>
                    <button class="btn btn-danger btn-sm" @click="handleDelete(target.id)">
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="card p-5">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ selectedTarget ? t('admin.modelStatus.historyTitle', { name: selectedTarget.name }) : t('admin.modelStatus.history') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.modelStatus.historyHint') }}
            </p>
          </div>
          <button
            class="btn btn-secondary"
            :disabled="!selectedTarget"
            @click="selectedTarget && loadChecks(selectedTarget.id)"
          >
            {{ t('common.refresh') }}
          </button>
        </div>

        <div v-if="!selectedTarget" class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
          {{ t('admin.modelStatus.selectTargetForHistory') }}
        </div>

        <div v-else-if="checksLoading" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.loading') }}
        </div>

        <div v-else-if="checks.length === 0" class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
          {{ t('admin.modelStatus.noHistory') }}
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="check in checks"
            :key="check.id"
            class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <span :class="statusBadgeClass(check.status)">
                  {{ statusLabel(check.status) }}
                </span>
                <span class="text-sm text-gray-600 dark:text-gray-300">
                  {{ check.latency_ms != null ? `${check.latency_ms} ms` : '-' }}
                </span>
              </div>
              <div class="text-xs text-gray-500 dark:text-gray-400">
                {{ formatDateTime(check.created_at) }}
              </div>
            </div>
            <div v-if="check.error_message" class="mt-3 whitespace-pre-wrap break-words rounded-xl bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-300">
              {{ check.error_message }}
            </div>
            <div v-else-if="check.response_text" class="mt-3 whitespace-pre-wrap break-words rounded-xl bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-dark-800 dark:text-gray-300">
              {{ check.response_text }}
            </div>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { accountsAPI } from '@/api/admin'
import modelStatusAPI, {
  type ModelStatusCheck,
  type ModelStatusOverview,
  type ModelStatusTarget
} from '@/api/admin/modelStatus'
import type { Account } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const submitting = ref(false)
const checksLoading = ref(false)
const runningId = ref<number | null>(null)
const includeDisabled = ref(true)

const overview = reactive<ModelStatusOverview>({
  total_targets: 0,
  enabled_targets: 0,
  healthy_targets: 0,
  failed_targets: 0,
  unknown_targets: 0,
  average_latency_ms: null,
  last_checked_target: null
})

const targets = ref<ModelStatusTarget[]>([])
const checks = ref<ModelStatusCheck[]>([])
const accounts = ref<Account[]>([])
const editingId = ref<number | null>(null)
const selectedTarget = ref<ModelStatusTarget | null>(null)

const form = reactive({
  name: '',
  account_id: 0,
  model_id: '',
  check_interval_seconds: 300,
  timeout_seconds: 45,
  enabled: true
})

const isEditing = computed(() => editingId.value !== null)
const accountOptions = computed(() => accounts.value.filter(account => account.id != null))
const healthyRatio = computed(() => {
  if (overview.total_targets === 0) return t('admin.modelStatus.summary.noTargets')
  return `${Math.round((overview.healthy_targets / overview.total_targets) * 100)}%`
})
const averageLatencyLabel = computed(() =>
  overview.average_latency_ms != null ? `${Math.round(overview.average_latency_ms)} ms` : '-'
)

watch(includeDisabled, () => {
  loadTargets()
  loadOverview()
})

onMounted(async () => {
  await Promise.all([loadAccounts(), refreshAll()])
})

function resetForm() {
  editingId.value = null
  form.name = ''
  form.account_id = 0
  form.model_id = ''
  form.check_interval_seconds = 300
  form.timeout_seconds = 45
  form.enabled = true
}

function startEdit(target: ModelStatusTarget) {
  editingId.value = target.id
  form.name = target.name
  form.account_id = target.account_id
  form.model_id = target.model_id
  form.check_interval_seconds = target.check_interval_seconds
  form.timeout_seconds = target.timeout_seconds
  form.enabled = target.enabled
}

function selectHistoryTarget(target: ModelStatusTarget) {
  selectedTarget.value = target
  void loadChecks(target.id)
}

async function refreshAll() {
  await Promise.all([loadTargets(), loadOverview()])
  if (selectedTarget.value) {
    await loadChecks(selectedTarget.value.id)
  }
}

async function loadOverview() {
  try {
    const data = await modelStatusAPI.getOverview(includeDisabled.value)
    Object.assign(overview, data)
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToLoad'))
  }
}

async function loadTargets() {
  loading.value = true
  try {
    targets.value = await modelStatusAPI.listTargets(includeDisabled.value)
    if (selectedTarget.value) {
      selectedTarget.value = targets.value.find(target => target.id === selectedTarget.value?.id) ?? null
    }
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToLoad'))
  } finally {
    loading.value = false
  }
}

async function loadAccounts() {
  try {
    const response = await accountsAPI.list(1, 200, { lite: 'true' })
    accounts.value = response.items
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToLoadAccounts'))
  }
}

async function loadChecks(targetId: number) {
  checksLoading.value = true
  try {
    checks.value = await modelStatusAPI.listChecks(targetId, 20)
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToLoadHistory'))
  } finally {
    checksLoading.value = false
  }
}

async function submitForm() {
  submitting.value = true
  const payload = {
    name: form.name.trim(),
    account_id: form.account_id,
    model_id: form.model_id.trim(),
    check_interval_seconds: form.check_interval_seconds,
    timeout_seconds: form.timeout_seconds,
    enabled: form.enabled
  }
  try {
    if (editingId.value) {
      await modelStatusAPI.updateTarget(editingId.value, payload)
      appStore.showSuccess(t('admin.modelStatus.updated'))
    } else {
      await modelStatusAPI.createTarget(payload)
      appStore.showSuccess(t('admin.modelStatus.created'))
    }
    resetForm()
    await refreshAll()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToSave'))
  } finally {
    submitting.value = false
  }
}

async function handleRun(id: number) {
  runningId.value = id
  try {
    await modelStatusAPI.runTarget(id)
    appStore.showSuccess(t('admin.modelStatus.runTriggered'))
    await refreshAll()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToRun'))
  } finally {
    runningId.value = null
  }
}

async function toggleEnabled(target: ModelStatusTarget) {
  try {
    await modelStatusAPI.updateTarget(target.id, { enabled: !target.enabled })
    appStore.showSuccess(target.enabled ? t('common.disabled') : t('common.enabled'))
    await refreshAll()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToSave'))
  }
}

async function handleDelete(id: number) {
  if (!window.confirm(t('admin.modelStatus.deleteConfirm'))) return
  try {
    await modelStatusAPI.deleteTarget(id)
    if (selectedTarget.value?.id === id) {
      selectedTarget.value = null
      checks.value = []
    }
    if (editingId.value === id) {
      resetForm()
    }
    appStore.showSuccess(t('admin.modelStatus.deleted'))
    await refreshAll()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.modelStatus.failedToDelete'))
  }
}

function statusLabel(status: ModelStatusTarget['latest_status'] | ModelStatusCheck['status']) {
  return t(`admin.modelStatus.status.${status}`)
}

function statusBadgeClass(status: ModelStatusTarget['latest_status'] | ModelStatusCheck['status']) {
  return [
    'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
    status === 'success'
      ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
      : status === 'failed'
        ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
        : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  ]
}
</script>
