<template>
  <div
    v-if="selectedIds.length > 0"
    class="mb-4 flex flex-col gap-3 rounded-lg bg-primary-50 p-3 dark:bg-primary-900/20 lg:flex-row lg:items-center lg:justify-between"
  >
    <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
      <span class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkActions.selected', { count: selectedIds.length }) }}
      </span>
      <button
        @click="$emit('select-page')"
        class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
      >
        {{ t('admin.accounts.bulkActions.selectCurrentPage') }}
      </button>
      <span class="text-gray-300 dark:text-primary-800">•</span>
      <button
        @click="$emit('invert-page')"
        class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
      >
        {{ t('admin.accounts.bulkActions.invertCurrentPage') }}
      </button>
      <span class="text-gray-300 dark:text-primary-800">•</span>
      <button
        @click="$emit('clear')"
        class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
      >
        {{ t('admin.accounts.bulkActions.clear') }}
      </button>
    </div>
    <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap">
      <button @click="$emit('delete')" class="btn btn-danger btn-sm">{{ t('admin.accounts.bulkActions.delete') }}</button>
      <button @click="$emit('test-connection')" class="btn btn-secondary btn-sm">{{ t('admin.accounts.bulkActions.testConnection') }}</button>
      <button @click="$emit('reset-status')" class="btn btn-secondary btn-sm">{{ t('admin.accounts.bulkActions.resetStatus') }}</button>
      <button @click="$emit('refresh-token')" class="btn btn-secondary btn-sm">{{ t('admin.accounts.bulkActions.refreshToken') }}</button>
      <button @click="$emit('toggle-schedulable', true)" class="btn btn-success btn-sm">{{ t('admin.accounts.bulkActions.enableScheduling') }}</button>
      <button @click="$emit('toggle-schedulable', false)" class="btn btn-warning btn-sm">{{ t('admin.accounts.bulkActions.disableScheduling') }}</button>
      <button @click="$emit('edit')" class="btn btn-primary btn-sm">{{ t('admin.accounts.bulkActions.edit') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps(['selectedIds'])
defineEmits([
  'delete',
  'edit',
  'clear',
  'select-page',
  'invert-page',
  'toggle-schedulable',
  'reset-status',
  'refresh-token',
  'test-connection'
])

const { t } = useI18n()
</script>
