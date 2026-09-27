<script lang="ts" setup>
import { computed } from 'vue';
import { useSongbeamerStore, type PageItem } from '../stores/songbeamer';
import PageContent from '../components/PageContent.vue';

const sb = useSongbeamerStore();

const page = computed(() => sb.presentation.pages[sb.presentation.page - 1] as PageItem | undefined);
</script>

<template>
    <div class="h-full w-full">
        <PageContent :page="page" aspect-ratio="auto" class="h-full w-full" />
    </div>
</template>
