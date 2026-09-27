<script setup lang="ts">
import { useConfigStore } from '../stores/config';
import { PlusIcon, MinusIcon } from '@lucide/vue';
import { clamp } from '../utils';
import { useSongbeamerStore } from '../stores/songbeamer';
import { watch } from 'vue';
import { SUPPORTED_ENCODINGS } from '../utils/constants';

const config = useConfigStore();
const sb = useSongbeamerStore();
const addGridSize = (s = 1) => (config.presentation.gridSize = clamp(config.presentation.gridSize + s, 1, 10));
const addGridCols = (s = 1) => (config.presentation.gridCaptureCols = clamp(config.presentation.gridCaptureCols + s, 1, 10));

watch(
    () => [config.presentation.localPdfRendering, config.presentation.localSngRendering],
    () => {
        sb.forceResync();
    }
);

const hasWakeLockSupport = !!navigator?.wakeLock;

function reset() {
    config.reset();
    location.reload();
}
</script>

<template>
    <div class="border-base-300 bg-base-200 mx-auto rounded-md border px-4 py-4 shadow-md lg:my-4 lg:w-4xl">
        <h1 class="font-heading">Settings</h1>

        <fieldset class="fieldset border-base-300 rounded-box flex w-full flex-wrap gap-4 border p-4">
            <legend class="fieldset-legend">Theme</legend>

            <label class="flex cursor-pointer items-center gap-2">
                <input type="radio" class="radio radio-sm" value="default" v-model="config.general.theme" />
                System
            </label>
            <label class="flex cursor-pointer items-center gap-2">
                <input type="radio" class="radio radio-sm" value="light" v-model="config.general.theme" />
                Light
            </label>
            <label class="flex cursor-pointer items-center gap-2">
                <input type="radio" class="radio radio-sm" value="dracula" v-model="config.general.theme" />
                Dark
            </label>
            <label class="flex cursor-pointer items-center gap-2">
                <input type="radio" class="radio radio-sm" value="cupcake" v-model="config.general.theme" />
                Cupcake
            </label>
            <label class="flex cursor-pointer items-center gap-2">
                <input type="radio" class="radio radio-sm" value="coffee" v-model="config.general.theme" />
                Coffee
            </label>
        </fieldset>

        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid w-full gap-4 border p-4">
            <legend class="fieldset-legend">General</legend>

            <label class="label">Keep screen awake</label>
            <div>
                <input type="checkbox" :disabled="!hasWakeLockSupport" v-model="config.general.wakeLock" class="toggle" />
                <div class="badge badge-error badge-xs ml-2" v-if="!hasWakeLockSupport">Not supported by browser</div>
            </div>

            <label class="label">Show Sidebar</label>
            <input type="checkbox" v-model="config.layout.showSidebar" class="toggle" />

            <label class="label">Show Navigation Menu</label>
            <div class="">
                <input type="checkbox" v-model="config.layout.showNavigation" class="toggle" />
                <span class="badge badge-accent badge-xs ml-2">Navigation by URL will always work</span>
            </div>

            <label class="label">Playlist Coloring</label>
            <select class="select select-sm w-32" v-model="config.layout.playlistColorMode">
                <option>off</option>
                <option>bar</option>
                <option>text</option>
            </select>
        </fieldset>

        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid gap-4 border p-4">
            <legend class="fieldset-legend">Rendering</legend>

            <label class="label"> Render PDF locally </label>
            <div>
                <input type="checkbox" v-model="config.presentation.localPdfRendering" class="toggle" />
                <div class="badge badge-accent badge-xs ml-2">Recomended</div>
            </div>

            <label class="label"> Render SNG as text </label>
            <div>
                <input type="checkbox" v-model="config.presentation.localSngRendering" class="toggle" />
                <div class="badge badge-accent badge-xs ml-2">Recomended</div>
            </div>

            <label class="label" v-if="config.presentation.localSngRendering">SNG Encoding</label>
            <select class="select select-sm w-32" v-model="config.presentation.sngEncoding" v-if="config.presentation.localSngRendering">
                <option v-for="enc in SUPPORTED_ENCODINGS">{{ enc }}</option>
            </select>

            <label class="label"> Use Grid Splitting </label>
            <div class="flex items-center gap-3 flex-wrap">
                <input type="checkbox" v-model="config.presentation.gridCaptureEnabled" class="toggle" />
                <div class="join" :class="{'*:btn-disabled': !config.presentation.gridCaptureEnabled}">
                    <button class="btn join-item btn-square btn-sm" @click="addGridCols(-1)">
                        <MinusIcon class="h-3 w-3" />
                    </button>
                    <button class="btn join-item btn-square btn-sm">{{ config.presentation.gridCaptureCols }}</button>
                    <button class="btn join-item btn-square btn-sm" @click="addGridCols(1)">
                        <PlusIcon class="h-3 w-3" />
                    </button>
                </div>
            </div>
            <p class="text-xs text-base-content/30 col-span-2 wrap-normal">
                If you enable grid splitting, make sure to set the number of columns to match the number of columns in songbeamer. Otherwise, the grid will be split incorrectly and the output will look wrong.
            </p>
        </fieldset>

        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid gap-4 border p-4">
            <legend class="fieldset-legend">GridView</legend>

            <label class="label">Grid Size</label>
            <div class="join">
                <button class="btn join-item btn-square btn-sm" @click="addGridSize(-1)">
                    <MinusIcon class="h-3 w-3" />
                </button>
                <button class="btn join-item btn-square btn-sm">{{ config.presentation.gridSize }}</button>
                <button class="btn join-item btn-square btn-sm" @click="addGridSize(1)">
                    <PlusIcon class="h-3 w-3" />
                </button>
            </div>

            <label class="label">Show Playlist Controls</label>
            <input type="checkbox" v-model="config.layout.showPlaylistCtrl" class="toggle" />

            <label class="label">Show Page Header</label>
            <input type="checkbox" v-model="config.layout.showGridPageHeader" class="toggle" />
        </fieldset>

        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid gap-4 border p-4">
            <legend class="fieldset-legend">RemoteView</legend>

            <label class="label">Show Playlist Controls</label>
            <input type="checkbox" v-model="config.layout.showRemotePlaylistControls" class="toggle" />
        </fieldset>

        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid gap-4 border p-4">
            <legend class="fieldset-legend">Header Items</legend>

            <label class="label">Show Black State Button</label>
            <input type="checkbox" v-model="config.layout.showStateButtonBlack" class="toggle" />

            <label class="label">Show Background State Button</label>
            <input type="checkbox" v-model="config.layout.showStateButtonBackground" class="toggle" />

            <label class="label">Show Page State Button</label>
            <input type="checkbox" v-model="config.layout.showStateButtonPage" class="toggle" />

            <label class="label">Show Logo State Button</label>
            <input type="checkbox" v-model="config.layout.showStateButtonLogo" class="toggle" />

            <label class="label">Show Language Switcher</label>
            <input type="checkbox" v-model="config.layout.showLanguageSelector" class="toggle" />
        </fieldset>
        <fieldset class="fieldset border-base-300 rounded-box fieldset-grid-cols grid gap-4 border p-4">
            <legend class="fieldset-legend">Footer</legend>

            <label class="label">Show Connection State</label>
            <input type="checkbox" v-model="config.layout.showConnectionStatus" class="toggle" />

            <label class="label">Show Fullscreen</label>
            <input type="checkbox" v-model="config.layout.showFullscreenButton" class="toggle" />

            <label class="label">Show Message</label>
            <input type="checkbox" v-model="config.layout.showMessageButton" class="toggle" />

            <label class="label">Show Grid Size Control</label>
            <input type="checkbox" v-model="config.layout.showGridSizeCtrl" class="toggle" />

            <label class="label">Show Video</label>
            <input type="checkbox" v-model="config.layout.showVideoCtrl" class="toggle" />

            <label class="label">Show Live Video</label>
            <input type="checkbox" v-model="config.layout.showLiveVideo" class="toggle" />

            <label class="label">Show Clock</label>
            <input type="checkbox" v-model="config.layout.showClock" class="toggle" />
        </fieldset>

        <button class="btn btn-sm btn-error m-auto mt-5" @click="reset()">Reset Settings</button>
    </div>
</template>

<style scoped>
.fieldset-grid-cols {
    grid-template-columns: min-content 1fr;
}
</style>
