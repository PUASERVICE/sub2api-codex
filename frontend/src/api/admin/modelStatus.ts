import { apiClient } from '../client'

export interface ModelStatusTarget {
  id: number
  name: string
  account_id: number
  account_name: string
  account_platform: string
  account_status: string
  model_id: string
  check_interval_seconds: number
  timeout_seconds: number
  enabled: boolean
  latest_status: 'success' | 'failed' | 'unknown'
  latest_latency_ms: number | null
  latest_error_message: string
  latest_response_text: string
  consecutive_failures: number
  last_checked_at: string | null
  last_success_at: string | null
  last_failure_at: string | null
  next_check_at: string | null
  created_at: string
  updated_at: string
}

export interface ModelStatusCheck {
  id: number
  target_id: number
  status: 'success' | 'failed' | 'unknown'
  latency_ms: number | null
  error_message: string
  response_text: string
  started_at: string
  finished_at: string
  created_at: string
}

export interface ModelStatusOverview {
  total_targets: number
  enabled_targets: number
  healthy_targets: number
  failed_targets: number
  unknown_targets: number
  average_latency_ms: number | null
  last_checked_target: string | null
}

export interface CreateModelStatusTargetRequest {
  name?: string
  account_id: number
  model_id: string
  check_interval_seconds?: number
  timeout_seconds?: number
  enabled?: boolean
}

export interface UpdateModelStatusTargetRequest {
  name?: string
  account_id?: number
  model_id?: string
  check_interval_seconds?: number
  timeout_seconds?: number
  enabled?: boolean
}

export async function getOverview(includeDisabled: boolean = true): Promise<ModelStatusOverview> {
  const { data } = await apiClient.get<ModelStatusOverview>('/admin/model-status/overview', {
    params: { include_disabled: includeDisabled }
  })
  return data
}

export async function listTargets(includeDisabled: boolean = true): Promise<ModelStatusTarget[]> {
  const { data } = await apiClient.get<ModelStatusTarget[]>('/admin/model-status/targets', {
    params: { include_disabled: includeDisabled }
  })
  return data ?? []
}

export async function createTarget(payload: CreateModelStatusTargetRequest): Promise<ModelStatusTarget> {
  const { data } = await apiClient.post<ModelStatusTarget>('/admin/model-status/targets', payload)
  return data
}

export async function updateTarget(id: number, payload: UpdateModelStatusTargetRequest): Promise<ModelStatusTarget> {
  const { data } = await apiClient.put<ModelStatusTarget>(`/admin/model-status/targets/${id}`, payload)
  return data
}

export async function deleteTarget(id: number): Promise<void> {
  await apiClient.delete(`/admin/model-status/targets/${id}`)
}

export async function runTarget(id: number): Promise<ModelStatusTarget> {
  const { data } = await apiClient.post<ModelStatusTarget>(`/admin/model-status/targets/${id}/run`)
  return data
}

export async function listChecks(id: number, limit: number = 20): Promise<ModelStatusCheck[]> {
  const { data } = await apiClient.get<ModelStatusCheck[]>(`/admin/model-status/targets/${id}/checks`, {
    params: { limit }
  })
  return data ?? []
}

export const modelStatusAPI = {
  getOverview,
  listTargets,
  createTarget,
  updateTarget,
  deleteTarget,
  runTarget,
  listChecks
}

export default modelStatusAPI
