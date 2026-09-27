<script lang="ts" setup>
import { useSongbeamerStore } from '../stores/songbeamer';
import IconLogo from '../assets/icons/IconLogo.vue';
import { useConfigStore } from '../stores/config';
import { sbColorToHex } from '../utils/color';

const sb = useSongbeamerStore();
const config = useConfigStore();
</script>

<template>
    <div class="drawer-side border-base-300 h-full border-r">
        <label for="mainDrawer" aria-label="Close sidebar" class="drawer-overlay"></label>
        <div class="bg-base-200 flex h-full w-10/12 flex-col sm:w-72">
            <div class="flex shrink-0 items-center gap-3 p-2">
                <button class="btn btn-square btn-accent btn-lg">
                    <IconLogo />
                </button>
                <div class="text-base-content">
                    <h1 class="text-lg/tight font-bold">Webremote</h1>
                    <p class="text-sm/tight">for Songbeamer</p>
                </div>
            </div>
            <div class="grow overflow-y-auto">
                <ul class="menu">
                    <li class="menu-title">Playlist</li>
                    <li v-for="(item, i) in sb.playlist.items" :style="{ '--item-color': sbColorToHex(item.color) }">
                        <span class="menu-title" v-if="!!item.caption.match(/^\[\[(.+)\]\]$/)">{{ item.caption.replace(/^\[\[(.+)\]\]$/, '$1') }}</span>
                        <a
                            :class="{
                                'menu-active': sb.playlist.itemindex === i,
                                'text-contrast-outline font-bold text-(--item-color)': config.layout.playlistColorMode === 'text'
                            }"
                            class="relative overflow-hidden"
                            @click="sb.setPlaylistItemindex(i)"
                            v-else
                        >
                            {{ item.caption }}
                            <div class="absolute left-0 h-full w-1" style="background-color: var(--item-color)" v-if="config.layout.playlistColorMode === 'bar'"></div>
                        </a>
                    </li>
                </ul>
            </div>
            <div class="shrink-0"></div>
        </div>
    </div>
</template>
