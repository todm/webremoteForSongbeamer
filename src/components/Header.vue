<script lang="ts" setup>
import {
    PanelLeftIcon,
    MonitorXIcon,
    WallpaperIcon,
    MonitorCheckIcon,
    MonitorCloudIcon,
    LanguagesIcon,
    GlobeIcon,
    LayoutGridIcon,
    SettingsIcon,
    MenuIcon,
    GalleryThumbnailsIcon,
    GamepadDirectionalIcon
} from '@lucide/vue';
import { useConfigStore } from '../stores/config';
import { useSongbeamerStore, PresentationState } from '../stores/songbeamer';

const isSidebarOpen = defineModel<boolean>();
const sb = useSongbeamerStore();
const config = useConfigStore();

const BTN_I = 'btn-accent';
const BTN_O = 'btn-active';

function changeLang(e: Event, lang: string) {
    e.preventDefault();
    e.stopPropagation();
    const start = Date.now();
    e.target?.addEventListener(
        e.type === 'mousedown' ? 'mouseup' : 'touchend',
        () => {
            if (lang === '') return sb.setLanguages('');
            if (Date.now() - start < 500) return sb.setLanguages(lang);
            sb.setLanguages(sb.presentation.languages.includes(lang) ? sb.presentation.languages.replace(lang, '') : sb.presentation.languages + lang);
        },
        { once: true }
    );
}
</script>

<template>
    <header class="bg-base-200 border-base-300 flex items-center gap-2 border-b p-2">
        <button class="btn btn-ghost btn-square btn-sm" @click="isSidebarOpen = !isSidebarOpen" v-if="config.layout.showSidebar"><PanelLeftIcon class="h-4 w-4" /></button>

        <button class="btn btn-sm" popovertarget="popover-1" style="anchor-name: --anchor-1" v-if="config.layout.showNavigation">
            <MenuIcon class="h-3 w-3" />
            <span class="hidden md:inline">{{ $route.name }}</span>
        </button>
        <ul class="dropdown menu rounded-box bg-base-100 w-52 shadow-sm" popover id="popover-1" style="position-anchor: --anchor-1">
            <li class="">
                <RouterLink to="/" exact-active-class="bg-base-300"><LayoutGridIcon class="h-4 w-4" />GridView</RouterLink>
            </li>
            <!-- <li>
                <RouterLink to="/preview" exact-active-class="bg-base-300"><IconEye class="h-4 w-4" />Preview</RouterLink>
            </li> -->
            <li>
                <RouterLink to="/presenter" exact-active-class="bg-base-300"><GalleryThumbnailsIcon class="h-4 w-4" />Presenter</RouterLink>
            </li>
            <li>
                <RouterLink to="/remote" exact-active-class="bg-base-300"><GamepadDirectionalIcon class="h-4 w-4" />Remote</RouterLink>
            </li>
            <li>
                <RouterLink to="/settings" exact-active-class="bg-base-300"><SettingsIcon class="h-4 w-4" />Settings</RouterLink>
            </li>
        </ul>

        <div class="grow"></div>
        <div class="join">
            <button
                v-if="config.layout.showStateButtonBlack"
                class="btn btn-square join-item"
                :class="sb.presentation.state === PresentationState.Black ? BTN_I : BTN_O"
                @click="sb.setPresentationState(PresentationState.Black)"
            >
                <MonitorXIcon class="h-4 w-4" />
            </button>
            <button
                v-if="config.layout.showStateButtonBackground"
                class="btn btn-square join-item"
                :class="sb.presentation.state === PresentationState.Background ? BTN_I : BTN_O"
                @click="sb.setPresentationState(PresentationState.Background)"
            >
                <WallpaperIcon class="h-4 w-4" />
            </button>
            <button
                v-if="config.layout.showStateButtonPage"
                class="btn btn-square join-item"
                :class="sb.presentation.state === PresentationState.Page ? BTN_I : BTN_O"
                @click="sb.setPresentationState(PresentationState.Page)"
            >
                <MonitorCheckIcon class="h-4 w-4" />
            </button>
            <button
                v-if="config.layout.showStateButtonLogo"
                class="btn btn-square join-item"
                :class="sb.presentation.state === PresentationState.Logo ? BTN_I : BTN_O"
                @click="sb.setPresentationState(PresentationState.Logo)"
            >
                <MonitorCloudIcon class="h-4 w-4" />
            </button>
        </div>
        <div class="relative" v-if="config.layout.showLanguageSelector">
            <button class="btn invisible"></button>
            <div class="fab absolute top-0 right-0 z-auto flex-col">
                <div tabindex="0" role="button" class="btn btn-square relative overflow-hidden">
                    <LanguagesIcon class="h-4 w-4" />
                    <div class="absolute inset-0 grid grid-cols-2 gap-0.5">
                        <div class="rounded-sm" :class="sb.presentation.languages.includes('1') ? 'bg-accent/50' : ''"></div>
                        <div class="rounded-sm" :class="sb.presentation.languages.includes('2') ? 'bg-accent/50' : ''"></div>
                        <div class="rounded-sm" :class="sb.presentation.languages.includes('3') ? 'bg-accent/50' : ''"></div>
                        <div class="rounded-sm" :class="sb.presentation.languages.includes('4') ? 'bg-accent/50' : ''"></div>
                    </div>
                </div>

                <button class="btn btn-square" @mousedown="changeLang($event, '')" @touchstart="changeLang($event, '')" :class="sb.presentation.languages === '' ? BTN_I : BTN_O">
                    <GlobeIcon class="h-4 w-4" />
                </button>
                <button
                    class="btn btn-square"
                    @mousedown="changeLang($event, '1')"
                    @touchstart="changeLang($event, '1')"
                    :class="sb.presentation.languages.includes('1') ? BTN_I : BTN_O"
                >
                    1
                </button>
                <button
                    class="btn btn-square"
                    @mousedown="changeLang($event, '2')"
                    @touchstart="changeLang($event, '2')"
                    :class="sb.presentation.languages.includes('2') ? BTN_I : BTN_O"
                >
                    2
                </button>
                <button
                    class="btn btn-square"
                    @mousedown="changeLang($event, '3')"
                    @touchstart="changeLang($event, '3')"
                    :class="sb.presentation.languages.includes('3') ? BTN_I : BTN_O"
                >
                    3
                </button>
                <button
                    class="btn btn-square"
                    @mousedown="changeLang($event, '4')"
                    @touchstart="changeLang($event, '4')"
                    :class="sb.presentation.languages.includes('4') ? BTN_I : BTN_O"
                >
                    4
                </button>
            </div>
        </div>
    </header>
</template>
