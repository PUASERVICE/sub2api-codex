import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import AccountBulkActionsBar from '../AccountBulkActionsBar.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'admin.accounts.bulkActions.selected') {
          return `${params?.count ?? 0} selected`
        }
        return key
      }
    })
  }
})

describe('AccountBulkActionsBar', () => {
  it('展示本页反选按钮并正确触发事件', async () => {
    const wrapper = mount(AccountBulkActionsBar, {
      props: {
        selectedIds: [1, 2]
      }
    })

    const invertButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'admin.accounts.bulkActions.invertCurrentPage')

    expect(invertButton).toBeDefined()

    await invertButton!.trigger('click')

    expect(wrapper.emitted('invert-page')).toHaveLength(1)
  })
})
