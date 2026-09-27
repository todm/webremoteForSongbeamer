import { defineStore } from 'pinia';
import { reactive, toRefs } from 'vue';
import { XOscClient } from '../services/OscClient';
import { setValueAtPath } from '../utils';
import { wait } from '../utils';
import { getPDFThumbnails, getSngThumbnails, getSongbeamerThumbnails } from '../services/Thumbnails';
import { useConfigStore } from './config';

export const useSongbeamerStore = defineStore('songbeamer', () => {
    const state = reactive({
        presentation: {
            state: PresentationState.Black as PresentationState,
            page: 0,
            pagecount: 0,
            pagecaption: '',
            filename: '',
            primarylanguage: 0,
            languages: '',
            message: {
                text: '',
                visible: IntBool.False as IntBool
            },
            permanentblack: IntBool.False as IntBool,
            pages: [] as PageItem[],
        },
        stage: {},
        playlist: {
            itemindex: 0,
            filename: '',
            count: 0,
            items: [] as PlaylistItem[],
            changed: 0 //Sync only
        },
        video: {
            state: VideoState.Stop as VideoState,
            position: 0,
            length: 0,
            filename: ''
        },
        livevideo: {
            state: VideoState.Stop as VideoState
        }
    });

    const url = new URL('/osc', location.href);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';

    const osc = new XOscClient(url.toString());
    osc.on('message', _onMessage);
    osc.connect();
    osc.sync(state, true);

    const config = useConfigStore();
    const cache = new Map<string, any[]>();

    let lastMessage = Date.now();
    function _onMessage(path: string, args: any[]) {
        lastMessage = Date.now();
        setValueAtPath(state, path, args[0]);
        switch (path) {
            case '/playlist/changed':
                _syncPlaylist();
                break;
            case '/playlist/itemindex':
                _syncPages();
                break;
        }
    }

    function _syncPlaylist() {
        state.playlist.items = Array(state.playlist.count)
            .fill(0)
            .map(() => ({ caption: '', filename: '', streamclass: 0, color: 0 }));
        osc.sync(state, true, '/playlist/items/');
    }

    async function _syncPages() {
        state.presentation.pages = Array(state.presentation.pagecount)
            .fill(0)
            .map(_ => ({}));

        const ext = state.presentation.filename.split('.').pop()?.toLowerCase() || '';
        const cacheKey = _getCurrentCacheKey();
        if (ext === 'pdf' && config.presentation.localPdfRendering) {
            const pages = cache.get(cacheKey) || (await getPDFThumbnails(1));
            pages.forEach((e, i) => (state.presentation.pages[i]._img = e));
            cache.set(cacheKey, pages);
        } else if (ext === 'sng' && config.presentation.localSngRendering) {
            const pages = cache.get(cacheKey) || (await getSngThumbnails(state.presentation.languages === '', config.presentation.sngEncoding || 'utf-8'));
            pages.forEach((e, i) => (state.presentation.pages[i]._sng = e));
            cache.set(cacheKey, pages);
        } else {
            const pages =
                cache.get(cacheKey) ||
                (await getSongbeamerThumbnails(state.presentation.pagecount, config.presentation.gridCaptureEnabled ? config.presentation.gridCaptureCols : undefined));
            pages.forEach((e, i) => (state.presentation.pages[i]._img = e));
            cache.set(cacheKey, pages);
        }
    }

    function _getCurrentCacheKey() {
        return `${state.presentation.filename};${state.playlist.itemindex};${state.presentation.languages}`;
    }

    function forceResync() {
        clearCache();
        osc.sync(state, true);
    }

    function clearCache() {
        cache.clear();
    }

    function setPlaylistItemindex(index: number) {
        osc.send('/playlist/itemindex', index);
    }

    function forwardPlaylistItemindex() {
        if (state.playlist.itemindex + 1 >= state.playlist.count) return;
        setPlaylistItemindex(state.playlist.itemindex + 1);
    }

    function backwardPlaylistItemindex() {
        if (state.playlist.itemindex <= 0) return;
        setPlaylistItemindex(state.playlist.itemindex - 1);
    }

    function setPresentationPage(page: number, activate?: boolean) {
        osc.send('/presentation/page', page);
        if (activate) setPresentationState(PresentationState.Page);
    }

    function forwardPresentationPage() {
        if (state.presentation.page + 1 > state.presentation.pagecount) return;
        setPresentationPage(state.presentation.page + 1);
    }

    function backwardPresentationPage() {
        if (state.presentation.page <= 1) return;
        setPresentationPage(state.presentation.page - 1);
    }

    function setPresentationState(state: PresentationState) {
        osc.send('/presentation/state', state);
    }

    function setLanguages(languages: string) {
        osc.send('/presentation/languages', languages);
    }

    function setVideoState(state: VideoState) {
        osc.send('/video/state', state);
    }

    function setVideoPosition(position: number) {
        osc.send('/video/position', position);
    }

    function setLiveVideoState(state: VideoState) {
        osc.send('/livevideo/state', state);
    }

    function setPresentationMessageText(text: string, show?: boolean) {
        osc.send('/presentation/message/text', text);
        if (show) setPresentationMessageVisible(true);
    }

    function setPresentationMessageVisible(visible: boolean) {
        osc.send('/presentation/message/visible', visible ? IntBool.True : IntBool.False);
    }

    async function getStatus() {
        const status = osc.status();
        if (status === WebSocket.CLOSED) return 'Offline';
        if (status === WebSocket.CONNECTING) return 'Offline';
        if (Date.now() - lastMessage < 1000) return 'Connected';

        osc.send('/info');
        await wait(250);
        if (Date.now() - lastMessage < 1000) return 'Connected';
        else return 'Connecting';
    }

    return {
        ...toRefs(state),
        setPlaylistItemindex,
        setPresentationPage,
        setPresentationState,
        setLanguages,
        forwardPlaylistItemindex,
        backwardPlaylistItemindex,
        forwardPresentationPage,
        backwardPresentationPage,
        getStatus,
        setVideoState,
        setVideoPosition,
        setLiveVideoState,
        setPresentationMessageText,
        setPresentationMessageVisible,
        forceResync,
        clearCache
    };
});

export const PresentationState = {
    Black: 0,
    Background: 1,
    Page: 2,
    Logo: 3
} as const;
export type PresentationState = (typeof PresentationState)[keyof typeof PresentationState];

export const IntBool = {
    False: 0,
    True: 1
} as const;
export type IntBool = (typeof IntBool)[keyof typeof IntBool];

export const VideoState = {
    Stop: 0,
    Play: 1,
    Pause: 2
} as const;
export type VideoState = (typeof VideoState)[keyof typeof VideoState];

export type PlaylistItem = {
    caption: string;
    filename: string;
    streamclass: number;
    color?: any;
};

export type PageItem = {
    caption?: string;
    preview?: string;
    _img?: string;
    _text?: string;
    _sng?: any;
};
