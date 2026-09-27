<script lang="ts" setup>
import { computed } from 'vue';
import PageContent from '../components/PageContent.vue';
import { PresentationState, useSongbeamerStore, type PageItem } from '../stores/songbeamer';
import { ChevronLeftIcon, ChevronRightIcon } from '@lucide/vue';

const sb = useSongbeamerStore();

const page = computed(() => sb.presentation.pages[sb.presentation.page - 1] as PageItem | undefined);
</script>

<template>
    <div class="flex h-full w-full flex-col gap-2 p-2">
        <h1 class="font-heading">{{ sb.playlist.items[sb.playlist.itemindex]?.caption }}</h1>
        <div class="flex h-20 grow gap-2">
            <button class="btn hidden h-full lg:flex w-1/12" @click="sb.backwardPresentationPage()"><ChevronLeftIcon /></button>
            <PageContent class="h-full w-full grow" :page="page" aspect-ratio="auto" />
            <button class="btn hidden h-full lg:flex w-1/12" @click="sb.forwardPresentationPage()"><ChevronRightIcon /></button>
        </div>
        <div class="flex h-20 gap-2 lg:hidden">
            <button class="btn h-full grow" @click="sb.backwardPresentationPage()"><ChevronLeftIcon /></button>
            <button class="btn h-full grow" @click="sb.forwardPresentationPage()"><ChevronRightIcon /></button>
        </div>
        <div class="flex h-3/12 gap-2 overflow-x-auto p-2">
            <div
                v-for="(page, i) in sb.presentation.pages"
                class="bg-base-200 shrink-0 cursor-pointer overflow-hidden rounded-md shadow-md outline-offset-1 select-none"
                :class="{
                    'outline-accent outline-3': sb.presentation.page === i + 1 && sb.presentation.state === PresentationState.Page,
                    'outline-accent/50 outline-3 outline-dashed': sb.presentation.page === i + 1 && sb.presentation.state !== PresentationState.Page
                }"
                @click="sb.setPresentationPage(i + 1, true)"
            >
                <PageContent :page="page" :index="i + 1" aspect-ratio="fixed" class="h-full" />
            </div>
        </div>
    </div>
</template>
