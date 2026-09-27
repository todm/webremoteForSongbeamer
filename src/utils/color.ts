export function sbColorToRGBA(color: number) {
    if (!color) return [0, 0, 0, 255];
    const a = new Uint8Array(4);
    new DataView(a.buffer).setUint32(0, color);
    return [...a];
}

export function sbColorToHex(color: number) {
    const a = sbColorToRGBA(color);
    return '#' + a.map(e => e.toString(16).padStart(2, '0')).join('');
}
