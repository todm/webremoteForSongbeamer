/**
 * Resolves after the specified number of milliseconds
 * @param ms
 * @returns Promise
 */
export const wait = (ms: number) => new Promise(r => setTimeout(r, ms));

/**
 * Gets the value at the specified path in the given object.
 * @param obj The object to retrieve the value from.
 * @param path The path to the value, with segments separated by '/'.
 * @returns The value at the specified path, or undefined if the path does not exist.
 */
export function getValueAtPath(obj: any, path: string): any {
    return path
        .split('/')
        .filter(Boolean)
        .reduce((e, p) => e?.[p], obj);
}

/**
 * Sets the value at the specified path in the given object, creating intermediate objects as needed.
 * @param obj The object to set the value in.
 * @param path The path to set the value at, with segments separated by '/'.
 * @param value The value to set at the specified path.
 * @returns The modified object with the value set at the specified path.
 */
export function setValueAtPath(obj: any, path: string, value: any): void {
    const parts = path.split('/').filter(Boolean);
    const lastPart = parts.pop();
    if (!lastPart) return;
    const target = parts.reduce((o, p) => (o[p] ??= {}), obj);
    target[lastPart] = value;
    return obj;
}

export function concatBytes(...byteArrays: Uint8Array[]) {
    const total = byteArrays.reduce((sum, a) => sum + a.length, 0);
    const out = new Uint8Array(total);
    let offset = 0;
    for (const arr of byteArrays) {
        out.set(arr, offset);
        offset += arr.length;
    }
    return out;
}

export function isObject(x: unknown) {
    return typeof x === 'object' && !Array.isArray(x) && x !== null;
}

export function ObjectAssignDeep(target: object, src: object) {
    if (!isObject(target) || !isObject(src)) throw new Error();
    const _target = target as any;
    for (const [key, val] of Object.entries(src)) {
        if (isObject(val)) {
            if (!isObject(_target[key])) _target[key] = {};
            ObjectAssignDeep(_target[key], val);
        } else {
            _target[key] = val;
        }
    }
    return target;
}

export function clamp(num: number, min: number, max: number) {
    return Math.min(Math.max(num, min), max);
}

export function canvasToBlob(canvas: HTMLCanvasElement) {
    return new Promise<Blob>((resolve, reject) =>
        canvas.toBlob(b => {
            b ? resolve(b) : reject();
        }, 'image/webp')
    );
}

export function canvasToObjectUrl(canvas: HTMLCanvasElement) {
    return canvasToBlob(canvas).then(b => URL.createObjectURL(b));
}

export function loadImage(url: string, cachebuster?: boolean) {
    const _url = new URL(url, location.href);
    if (cachebuster) _url.searchParams.set('cache-buster', Math.random().toString(16).substring(2));
    return new Promise<HTMLImageElement>((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = err => reject(err);
        img.src = _url.toString();
    });
}

export async function cropImage(img: HTMLImageElement, x: number, y: number, w: number, h: number) {
    const canvas = document.createElement('canvas');
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext('2d', { alpha: true })!;
    ctx.drawImage(img, x, y);
    const url = await canvasToObjectUrl(canvas);
    return loadImage(url);
}


export { default as EventEmitter } from './EventEmitter';
