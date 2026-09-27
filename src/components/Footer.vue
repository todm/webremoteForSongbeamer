<script setup lang="ts">
import { ExpandIcon, FilmIcon, MessageSquareTextIcon, MinusIcon, PlusIcon, RefreshCwIcon, ShrinkIcon, CctvIcon } from '@lucide/vue';
import { useConfigStore } from '../stores/config';
import { clamp } from '../utils';
import { useFullscreen, useInterval } from '../utils/composables';
import { IntBool, useSongbeamerStore, VideoState } from '../stores/songbeamer';
import { ref } from 'vue';

const sb = useSongbeamerStore();
const config = useConfigStore();
const isVideoOpen = defineModel<boolean>('isVideoOpen');

const presentationMessageModal = ref<HTMLDialogElement>();

const addGridSize = (s = 1) => (config.presentation.gridSize = clamp(config.presentation.gridSize + s, 1, 10));

const statusText = ref('Offline');
const statusClass = ref('');
useInterval(
    async () => {
        const status = await sb.getStatus();
        statusText.value = status;
        if (status === 'Connected') statusClass.value = 'status-success';
        else if (status === 'Connecting') statusClass.value = 'status-warning animate-pulse';
        else statusClass.value = 'status-error';
    },
    2000,
    true
);

const currentTime = ref('00:00:00');
useInterval(
    () => {
        currentTime.value = new Date().toLocaleTimeString('en-US', { hour12: false });
    },
    1000,
    true
);

const { isFullscreen, toggleFullscreen } = useFullscreen();
</script>

<template>
    <footer class="bg-base-200 border-base-300 flex items-center gap-2 border-t p-2">
        <button class="btn btn-active btn-xs" v-if="config.layout.showConnectionStatus">
            <span class="status" :class="statusClass"></span>
            <span class="hidden md:inline">
                {{ statusText }}
            </span>
        </button>
        <button class="btn btn-xs btn-square" @click="sb.forceResync()">
            <RefreshCwIcon class="h-3 w-3" />
        </button>
        <span class="text-xs" v-if="config.layout.showClock">{{ currentTime }}</span>
        <div class="grow"></div>
        <button class="btn btn-xs" @click="isVideoOpen = !isVideoOpen" v-if="config.layout.showVideoCtrl" :class="{ 'btn-error': sb.video.state === VideoState.Play }">
            <FilmIcon class="h-3 w-3" />
            <span class="hidden lg:inline">Video</span>
        </button>
        <button
            class="btn btn-xs"
            @click="sb.setLiveVideoState(sb.livevideo.state === VideoState.Play ? VideoState.Stop : VideoState.Play)"
            :class="{ 'btn-error': sb.livevideo.state === VideoState.Play }"
            v-if="config.layout.showLiveVideo"
        >
            <CctvIcon class="h-4 w-4" />
            <span class="hidden lg:inline">Live</span>
        </button>
        <button
            class="btn btn-xs"
            @click="presentationMessageModal?.showModal()"
            v-if="config.layout.showMessageButton"
            :class="{ 'btn-info': sb.presentation.message.visible === IntBool.True }"
        >
            <MessageSquareTextIcon class="h-3 w-3" />
            <span class="hidden lg:inline">Message</span>
        </button>
        <div class="join" v-if="config.layout.showGridSizeCtrl">
            <button class="btn join-item btn-xs btn-square" @click="addGridSize(-1)">
                <MinusIcon class="h-3 w-3" />
            </button>
            <button class="btn join-item btn-xs btn-square hidden md:inline">{{ config.presentation.gridSize }}</button>
            <button class="btn join-item btn-xs btn-square" @click="addGridSize(1)">
                <PlusIcon class="h-3 w-3" />
            </button>
        </div>
        <button class="btn btn-xs" @click="toggleFullscreen()" v-if="config.layout.showFullscreenButton">
            <ExpandIcon class="h-3 w-3" v-if="!isFullscreen" />
            <ShrinkIcon class="h-3 w-3" v-else />
        </button>
    </footer>

    <dialog class="modal" ref="presentationMessageModal">
        <div class="modal-box">
            <h3 class="text-lg font-bold">Presentation Message</h3>
            <textarea class="textarea mt-4 w-full rounded-md" placeholder="Presentation Message" v-model="sb.presentation.message.text"></textarea>
            <div class="modal-action">
                <form method="dialog" class="space-x-2">
                    <button
                        class="btn btn-accent"
                        @click="
                            () => {
                                sb.setPresentationMessageText(sb.presentation.message.text, true);
                                presentationMessageModal?.close();
                            }
                        "
                    >
                        Show
                    </button>
                    <button
                        class="btn"
                        @click="
                            () => {
                                sb.setPresentationMessageVisible(false);
                                presentationMessageModal?.close();
                            }
                        "
                    >
                        Hide
                    </button>
                    <button class="btn" @click="presentationMessageModal?.close()">Cancel</button>
                </form>
            </div>
        </div>
        <form method="dialog" class="modal-backdrop">
            <button>close</button>
        </form>
    </dialog>
</template>
