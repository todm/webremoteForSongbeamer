<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useSongbeamerStore } from '../stores/songbeamer';
import { useInterval, useTimeout } from '../utils/composables';
import { useConfigStore } from '../stores/config';

const sb = useSongbeamerStore();
const cfg = useConfigStore();

const imgSrc = ref("")
const updateImgSrc = () => {
    imgSrc.value = `/sbfiles/preview?cacheBuster=${Date.now()}&q=${cfg.presentation.previewQuality}`
}
useInterval(updateImgSrc, 5000, true)

let queued = 0;
watch(() => ([
    sb.presentation.page,
    sb.presentation.languages,
    sb.presentation.state,
    sb.playlist.itemindex,
    sb.playlist.changed,
]), () => {
    queued++;
    useTimeout(() => {
        updateImgSrc()
        queued--;
    }, queued * 500 + 1);
})


</script>

<template>
    <div class="h-full w-full relative">
        <img :src="imgSrc" class="w-full h-full object-contain" @click="updateImgSrc">
        <div class="badge absolute bottom-2 right-2">Tap to update</div>
    </div>
</template>
