const TOKEN_KEY = 'poster_token'

export function getToken() {
  return sessionStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token) {
  if (token) sessionStorage.setItem(TOKEN_KEY, token)
  else sessionStorage.removeItem(TOKEN_KEY)
}

export async function request(path, options = {}) {
  const headers = new Headers(options.headers || {})
  const token = getToken()
  if (token) headers.set('Authorization', 'Bearer ' + token)
  let body = options.body
  if (options.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(options.json)
  }
  const response = await fetch(path, {
    method: options.method || 'POST',
    headers,
    body,
  })
  const type = response.headers.get('content-type') || ''
  if (!type.includes('application/json')) {
    if (!response.ok) {
      throw new Error('请求失败')
    }
    return response
  }
  const data = await response.json()
  if (typeof data.code === 'number' && data.code !== 200) {
    if (data.code === 401 && getToken()) {
      setToken('')
      if (!location.pathname.startsWith('/login') && !location.pathname.startsWith('/register')) {
        location.assign('/login')
      }
    }
    const error = new Error(data.msg || '请求失败')
    error.code = data.code
    throw error
  }
  return data
}

export function post(path, json) {
  return request(path, { json: json ?? {} })
}

export function upload(file) {
  const body = new FormData()
  body.append('file', file)
  return request('/api/asset/upload', { body })
}

export async function assetURL(id) {
  const response = await fetch('/api/asset/content/' + id, {
    headers: { Authorization: 'Bearer ' + getToken() },
  })
  const type = response.headers.get('content-type') || ''
  if (type.includes('application/json')) {
    const data = await response.json()
    throw new Error(data.msg || '素材无法打开')
  }
  const blob = await response.blob()
  return URL.createObjectURL(blob)
}
