<script setup lang="ts">
import { MessageSquareTextIcon, MoveLeftIcon, MoveRightIcon, XIcon } from '@lucide/vue';
import PageContent from '../components/PageContent.vue';
import { useConfigStore } from '../stores/config';
import { PresentationState, useSongbeamerStore } from '../stores/songbeamer';

const sb = useSongbeamerStore();
const config = useConfigStore();
</script>

<template>
    <div class="p-4">
        <div class="alert alert-info alert-soft mb-3" v-if="sb.presentation.message.visible">
            <MessageSquareTextIcon class="h-4 w-4" />
            <span>{{ sb.presentation.message.text }}</span>
            <button class="btn btn-xs btn-ghost" @click="sb.setPresentationMessageVisible(false)">
                <XIcon class="h-3 w-3" />
            </button>
        </div>
        <h1 class="font-heading">{{ sb.playlist.items[sb.playlist.itemindex]?.caption }}</h1>
        <div class="grid gap-2" :style="{ 'grid-template-columns': `repeat(${config.presentation.gridSize}, 1fr)` }">
            <div
                v-for="(page, i) in sb.presentation.pages"
                class="bg-base-200 cursor-pointer overflow-hidden rounded-md shadow-md outline-offset-1 select-none"
                :class="{
                    'outline-accent outline-3': sb.presentation.page === i + 1 && sb.presentation.state === PresentationState.Page,
                    'outline-accent/50 outline-3 outline-dashed': sb.presentation.page === i + 1 && sb.presentation.state !== PresentationState.Page
                }"
                @click="sb.setPresentationPage(i + 1, true)"
            >
                <PageContent :page="page" :index="i + 1" aspect-ratio="fixed" />
            </div>
        </div>
        <div class="join mt-5 flex justify-center" v-if="config.layout.showPlaylistCtrl">
            <button class="btn join-item" @click="sb.backwardPlaylistItemindex()"><MoveLeftIcon class="text-base-content/75" /></button>
            <button class="btn join-item" @click="sb.forwardPlaylistItemindex()"><MoveRightIcon class="text-base-content/75" /></button>
        </div>
    </div>
</template>
