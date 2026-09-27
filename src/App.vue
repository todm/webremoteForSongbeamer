<script lang="ts" setup>
import { watch } from 'vue';
import { useWakeLock } from './utils/composables';
import { useConfigStore } from './stores/config';

const config = useConfigStore();

watch(
    () => config.general.theme,
    () => {
        if (config.general.theme === 'default') document.documentElement.removeAttribute('data-theme');
        else document.documentElement.setAttribute('data-theme', config.general.theme);

        const themeColor = getComputedStyle(document.body).getPropertyValue('--color-base-300');
        document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')?.setAttribute('content', themeColor);
    },
    {
        immediate: true
    }
);

const wakeLock = useWakeLock(config.general.wakeLock);
watch(() => config.general.wakeLock, wakeLock.set);
</script>

<template>
    <router-view />
</template>
