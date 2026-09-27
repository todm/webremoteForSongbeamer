<script lang="ts" setup>
import { ChevronLeftIcon, ChevronRightIcon, MoveLeftIcon, MoveRightIcon } from '@lucide/vue';
import { useConfigStore } from '../stores/config';
import { useSongbeamerStore } from '../stores/songbeamer';

const sb = useSongbeamerStore();
const config = useConfigStore();
</script>

<template>
    <div class="layout m-auto h-full max-w-2xl gap-2 p-4">
        <button class="btn h-full" @click="sb.forwardPresentationPage()">
            <ChevronRightIcon class="h-10 w-10" />
        </button>
        <button class="btn h-full" @click="sb.backwardPresentationPage()">
            <ChevronLeftIcon class="h-10 w-10" />
        </button>
        <div class="join" v-if="config.layout.showRemotePlaylistControls">
            <button class="btn join-item h-full grow" @click="sb.backwardPlaylistItemindex()">
                <MoveLeftIcon class="h-10 w-10" />
            </button>
            <button class="btn join-item h-full grow" @click="sb.forwardPlaylistItemindex()">
                <MoveRightIcon class="h-10 w-10" />
            </button>
        </div>
        <div class="space-y-2">
            <div class="text-base-content/50 text-center">
                {{ sb.playlist.items[sb.playlist.itemindex]?.caption }}
            </div>
            <div class="text-base-content flex items-center justify-center gap-2 font-semibold">
                <span class="badge badge-info"> {{ sb.presentation.page }}/{{ sb.presentation.pagecount }} </span>
                {{ sb.presentation.pagecaption }}
            </div>
        </div>
    </div>
</template>

<style scoped>
.layout {
    display: grid;
    grid-template-rows: 5fr 2fr 1fr 1fr;
}
</style>
