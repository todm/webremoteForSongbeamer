<script setup lang="ts">
import { ref } from 'vue';
import Sidebar from '../components/Sidebar.vue';
import Header from '../components/Header.vue';
import Footer from '../components/Footer.vue';
import VideoBar from '../components/VideoBar.vue';
import { useConfigStore } from '../stores/config.ts';

const isSidebarOpen = ref(true);
const isVideoOpen = ref(false);
const config = useConfigStore();
</script>

<template>
    <div class="drawer h-full" :class="{ 'lg:drawer-open': isSidebarOpen }">
        <input id="mainDrawer" type="checkbox" class="drawer-toggle" v-model="isSidebarOpen" />
        <div class="drawer-content absolute flex h-full w-full flex-col">
            <Header v-model="isSidebarOpen" />
            <main class="grow overflow-auto">
                <router-view />
            </main>
            <VideoBar v-if="isVideoOpen && config.layout.showVideoCtrl" />
            <Footer v-model:is-video-open="isVideoOpen" />
        </div>
        <Sidebar v-if="config.layout.showSidebar" />
    </div>
</template>
