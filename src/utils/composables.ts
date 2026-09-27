import { onMounted, onUnmounted, ref } from 'vue';

export function useInterval(fn: () => void, ms: number, runImmediately?: boolean) {
    let interval: ReturnType<typeof setInterval>;
    onMounted(() => {
        interval = setInterval(fn, ms);
        if (runImmediately) fn();
    });
    onUnmounted(() => clearInterval(interval));
    return () => clearInterval(interval);
}

export function useEventListener(el: EventTarget, event: string, fn: (e: Event) => void) {
    onMounted(() => el.addEventListener(event, fn));
    onUnmounted(() => el.removeEventListener(event, fn));
    return () => el.removeEventListener(event, fn);
}

export function useFullscreen() {
    const isFullscreen = ref(!!document.fullscreenElement);
    const requestFullscreen = () => document.documentElement.requestFullscreen();
    const exitFullscreen = () => document.exitFullscreen();
    const toggleFullscreen = () => (isFullscreen.value ? exitFullscreen() : requestFullscreen());
    useEventListener(document, 'fullscreenchange', () => {
        isFullscreen.value = !!document.fullscreenElement;
    });
    return {
        requestFullscreen,
        exitFullscreen,
        toggleFullscreen,
        isFullscreen
    };
}

export function useWakeLock(onMount?: boolean) {
    const sentinel = ref<WakeLockSentinel>();
    const request = async () => {
        await release();
        sentinel.value = await navigator?.wakeLock?.request?.('screen');
    };
    const release = async () => {
        await sentinel.value?.release();
        sentinel.value = undefined;
    };
    const set = async (enabled: boolean) => {
        if (enabled) request();
        else release();
    };

    if (onMount) onMounted(() => request());
    onUnmounted(() => release());

    return {
        sentinel,
        set,
        request,
        release
    };
}
