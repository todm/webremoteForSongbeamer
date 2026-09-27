<script setup lang="ts">
import { computed } from 'vue';
import { PauseIcon, PlayIcon, SquareIcon } from '@lucide/vue';
import { useSongbeamerStore, VideoState } from '../stores/songbeamer';

const sb = useSongbeamerStore();
const SECONDS_DAY = 86400;

const positionStr = computed(() => {
    const secs = Math.floor(sb.video.position * SECONDS_DAY);
    const mm = Math.floor(secs / 60);
    const ss = secs % 60;
    return mm.toString().padStart(2, '0') + ':' + ss.toString().padStart(2, '0');
});

const lengthStr = computed(() => {
    const secs = Math.floor(sb.video.length * SECONDS_DAY);
    const mm = Math.floor(secs / 60);
    const ss = secs % 60;
    return mm.toString().padStart(2, '0') + ':' + ss.toString().padStart(2, '0');
});

const position = computed({
    get: () => sb.video.position * SECONDS_DAY,
    set: (val: number[]) => sb.setVideoPosition(+val / SECONDS_DAY)
});
</script>

<template>
    <section>
        <div class="bg-base-200 border-base-300 flex items-center gap-2 border-t border-b px-3 py-1">
            <span class="text-sm font-semibold">Video</span>
            <span class="text-base-content/50 grow overflow-hidden text-xs text-ellipsis whitespace-nowrap">{{ sb.video.filename }}</span>
        </div>
        <div class="flex flex-col gap-2 px-3 py-2">
            <div class="flex items-center justify-center gap-2">
                <button class="btn btn-md btn-square" @click="sb.setVideoState(VideoState.Play)" v-if="sb.video.state === VideoState.Pause || sb.video.state === VideoState.Stop">
                    <PlayIcon class="h-4 w-4" />
                </button>
                <button class="btn btn-md btn-square" @click="sb.setVideoState(VideoState.Pause)" v-else>
                    <PauseIcon class="h-4 w-4" />
                </button>
                <button class="btn btn-md btn-square" @click="sb.setVideoState(VideoState.Stop)"><SquareIcon class="h-4 w-4" /></button>
            </div>
            <input type="range" class="range range-xs w-full" v-model="position" min="0" :max="sb.video.length * SECONDS_DAY" />
            <div class="flex items-center gap-2">
                <span class="font-mono text-xs">{{ positionStr }}</span>
                <div class="grow"></div>
                <span class="font-mono text-xs">{{ lengthStr }}</span>
            </div>
        </div>
    </section>
</template>
