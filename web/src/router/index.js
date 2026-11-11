import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '../stores/session'
import StudioLayout from '../layouts/StudioLayout.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import HomeView from '../views/HomeView.vue'
import TemplateView from '../views/TemplateView.vue'
import PosterListView from '../views/PosterListView.vue'
import AssetView from '../views/AssetView.vue'
import EditorView from '../views/EditorView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginView, meta: { public: true } },
    { path: '/register', component: RegisterView, meta: { public: true } },
    {
      path: '/',
      component: StudioLayout,
      children: [
        { path: '', component: HomeView },
        { path: 'templates', component: TemplateView },
        { path: 'posters', component: PosterListView },
        { path: 'assets', component: AssetView },
      ],
    },
    { path: '/editor/:id', component: EditorView },
  ],
})

router.beforeEach((to) => {
  const session = useSession()
  if (!to.meta.public && !session.token) return '/login'
  if (to.meta.public && session.token) return '/'
})
