import { ref } from 'vue'
import { describe, expect, it } from 'vitest'

import { useTableSelection } from '@/composables/useTableSelection'

describe('useTableSelection', () => {
  it('selectVisible 会选中当前页所有数据', () => {
    const rows = ref([{ id: 1 }, { id: 2 }, { id: 3 }])
    const { selectedIds, selectVisible } = useTableSelection({
      rows,
      getId: (row) => row.id
    })

    selectVisible()

    expect(selectedIds.value).toEqual([1, 2, 3])
  })

  it('invertVisible 只反转当前页选中状态并保留其他页选择', () => {
    const rows = ref([{ id: 1 }, { id: 2 }])
    const { selectedIds, setSelectedIds, invertVisible } = useTableSelection({
      rows,
      getId: (row) => row.id
    })

    setSelectedIds([1, 3])
    invertVisible()

    expect(selectedIds.value).toEqual([3, 2])
  })
})
