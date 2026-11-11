import { defineStore } from 'pinia'
import { getToken, post, setToken } from '../api/http'

export const useSession = defineStore('session', {
  state: () => ({
    token: getToken(),
    account: null,
  }),
  actions: {
    accept(data) {
      this.token = data.token
      this.account = data.account
      setToken(data.token)
    },
    async register(form) {
      const result = await post('/api/account/register', form)
      this.accept(result.data)
    },
    async login(form) {
      const result = await post('/api/account/login', form)
      this.accept(result.data)
    },
    async refresh() {
      if (!this.token) return
      const result = await post('/api/account/one', {})
      this.account = result.data
    },
    async logout() {
      try {
        await post('/api/account/logout', {})
      } catch {
        /* 会话可能已经失效，本地仍要清掉。 */
      }
      this.token = ''
      this.account = null
      setToken('')
    },
  },
})
