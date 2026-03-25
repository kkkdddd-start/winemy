const DEFAULT_TIMEOUT = 30000

export interface WailsCallOptions {
  timeout?: number
  errorMessage?: string
}

export async function wailsCall<T>(
  fn: () => Promise<T>,
  options: WailsCallOptions = {}
): Promise<T> {
  const { timeout = DEFAULT_TIMEOUT, errorMessage = '操作超时' } = options

  const timeoutPromise = new Promise<never>((_, reject) => {
    setTimeout(() => {
      reject(new Error(errorMessage))
    }, timeout)
  })

  try {
    return await Promise.race([fn(), timeoutPromise])
  } catch (error) {
    if (error instanceof Error && error.message === errorMessage) {
      throw error
    }
    throw error
  }
}

export function createTimeoutErrorMessage(operation: string, timeout: number): string {
  return `${operation}失败: 请求超时 (${timeout}ms)`
}

export const WAILS_TIMEOUT = DEFAULT_TIMEOUT
export const API_TIMEOUT = 30000
export const LONG_OPERATION_TIMEOUT = 60000
