import { createRouter, createWebHistory } from 'vue-router'
import Projects from '../views/Projects.vue'
import Settings from '../views/Settings.vue'
import ProjectDetail from "@/views/ProjectDetail.vue";
import Identities from "@/views/Identities.vue";

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'dashboard',
            component: Projects
        },
        {
            path: '/projects/:id',
            name: 'project-detail',
            component: ProjectDetail
        },
        {
            path: '/settings',
            name: 'settings',
            component: Settings
        },
        {
            path: '/identities',
            name: 'identities',
            component: Identities
        },
        {
            path: '/organizations',
            name: 'organizations',
            component: () => import('../views/OrganizationList.vue')
        },
        {
            path: '/organizations/:id',
            name: 'organization-detail',
            component: () => import('../views/OrganizationDetail.vue')
        }
    ]
})

export default router
