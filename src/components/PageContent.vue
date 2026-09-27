<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useSongbeamerStore, type PageItem } from '../stores/songbeamer';
import { loadImage } from '../utils';
import { useConfigStore } from '../stores/config';

const sb = useSongbeamerStore();
const config = useConfigStore();

const { aspectRatio, page } = defineProps<{
    index?: number;
    page?: PageItem;
    aspectRatio?: number | 'auto' | 'fixed';
}>();

const ar = ref<'auto' | number>('auto');

watch(
    () => [page, page?._img],
    () => {
        if (page?._img && aspectRatio === 'fixed')
            loadImage(page?._img || '').then(img => {
                ar.value = img.naturalWidth / img.naturalHeight;
            });
        else if (typeof aspectRatio === 'number') ar.value = aspectRatio;
        else ar.value = 'auto';
    },
    { immediate: true }
);
</script>

<template>
    <div :style="{ aspectRatio: ar }">
        <div v-if="!!page?._img" class="h-full w-full bg-contain bg-center bg-no-repeat" :style="{ backgroundImage: `url('${page._img}')` }"></div>
        <div v-if="!!page?._sng" class="sb-html">
            <div class="bg-accent/50 px-4 py-1 font-bold" v-if="config.layout.showGridPageHeader">{{ index }} {{ page._sng.marker || ' ' }}</div>
            <div class="p-4">
                <p
                    v-for="p in page._sng.lines"
                    v-html="p.text"
                    v-show="sb.presentation.languages === '' || sb.presentation.languages.includes(p.language + 1)"
                    :class="{ 'opacity-50': p.language !== 0 && sb.presentation.languages.length !== 1 }"
                ></p>
            </div>
        </div>
        <div v-if="!!page?._text" class="">{{ page._text }}</div>
    </div>
</template>
