import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

import DashboardView from './DashboardView.vue';
import Grid from './Grid.vue';
import Settings from './Settings.vue';
import Preview from './Preview.vue';
import Presenter from './Presenter.vue';
import Remote from './Remote.vue';
import NotFound from './NotFound.vue';

const routes = [
    {
        name: 'main',
        path: '/',
        component: DashboardView,
        children: [
            { name: 'GridView', path: '/', component: Grid },
            { name: 'Preview', path: 'preview', component: Preview },
            { name: 'Presenter', path: 'presenter', component: Presenter },
            { name: 'Remote', path: 'remote', component: Remote },
            { name: 'Settings', path: 'settings', component: Settings }
        ]
    },
    { name: 'notfound', path: '/:pathMatch(.*)*', component: NotFound }
] as RouteRecordRaw[];

export const router = createRouter({
    history: createWebHistory(),
    routes
});
