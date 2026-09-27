import { defineStore } from 'pinia';
import { reactive, watch } from 'vue';
import { ObjectAssignDeep } from '../utils';

const localStorageKey = 'webremoteforsongbeamer-config';
export const useConfigStore = defineStore('config', () => {
    const state = reactive({
        presentation: {
            gridSize: 3,
            localPdfRendering: true,
            localSngRendering: true,
            sngEncoding: 'utf-8',
            gridCaptureEnabled: false,
            gridCaptureCols: 3
        },
        layout: {
            showNavigation: true,
            showSidebar: true,
            showLanguageSelector: true,
            showStateButtonBlack: true,
            showStateButtonBackground: true,
            showStateButtonPage: true,
            showStateButtonLogo: true,
            showPlaylistCtrl: true,
            showConnectionStatus: true,
            showFullscreenButton: true,
            showMessageButton: true,
            showGridSizeCtrl: true,
            showGridPageHeader: true,
            showVideoCtrl: true,
            showLiveVideo: true,
            showRemotePlaylistControls: true,
            showClock: true,
            playlistColorMode: 'off' as 'off' | 'bar' | 'text'
        },
        general: {
            theme: 'default',
            wakeLock: false
        }
    });

    function load() {
        const stored = JSON.parse(localStorage.getItem(localStorageKey) || '{}');
        ObjectAssignDeep(state, stored);
    }

    function save() {
        localStorage.setItem(localStorageKey, JSON.stringify(state));
    }

    function reset() {
        localStorage.removeItem(localStorageKey);
    }

    watch(state, save, { deep: true });
    load();
    save();

    return { ...state, reset, load, save };
});
