import { canvasToObjectUrl, cropImage, loadImage } from '../utils';

export async function getSongbeamerThumbnails(count: number, cols?: number) {
    let url = '/thumbnails';
    if (cols) url += `?cols=${cols}`;

    const img = await loadImage(url, true);
    const th = Math.ceil(img.naturalHeight / count);
    const images = await Promise.all([...Array(count).keys()].map(async i => cropImage(img, 0, -i * th, img.width, th)));
    return images.map(e => e.src);
}

export async function getPDFThumbnails(scale = 1) {
    const pdfjs = await import('pdfjs-dist');
    pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.js';

    const data = await fetch('/sbfiles/presentation', { cache: 'no-store' }).then(e => e.arrayBuffer());
    const pdf = await pdfjs.getDocument({ data }).promise;
    return Promise.all(
        [...Array(pdf.numPages).keys()].map(async i => {
            const page = await pdf.getPage(i + 1);
            const viewport = page.getViewport({ scale });
            const canvas = document.createElement('canvas');
            canvas.width = viewport.width;
            canvas.height = viewport.height;
            await page.render({
                canvas,
                canvasContext: canvas.getContext('2d', { alpha: true })!,
                viewport,
                intent: 'print'
            }).promise;
            return canvasToObjectUrl(canvas);
        })
    );
}

const VERSE_MARKERS = [
    'Unbekannt',
    'Unbenannt',
    'Unknown',
    'Intro',
    'Verse',
    'Vers',
    'Strophe',
    'Pre-Bridge',
    'Bridge',
    'Misc',
    'Pre-Refrain',
    'Refrain',
    'Pre-Chorus',
    'Chorus',
    'Pre-Coda',
    'Zwischenspiel',
    'Instrumental',
    'Interlude',
    'Coda',
    'Ending',
    'Outro',
    'Breakdown',
    'Vamp',
    'Turnaround',
    'Tag',
    'Andere',
    'Mittelteil',
    'Schluss-Chorus',
    'Schluss',
    'Gesprochen',
    'Oberstimme',
    'Post-Chorus',
    'Mid-Section',
    'Rap',
    'Spoken Words',
    'Ostinato Refrain',
    'Descant',
    'Hidden',
    'Invisible',
    'Comment',
    'Title',
    'Copyright'
] as const;

export async function getSngThumbnails(allLanguages?: boolean, encoding?: string) {
    const res = await fetch('/sbfiles/presentation', { cache: 'no-store' });
    const data = await res.arrayBuffer();

    encoding = new TextDecoder('utf-8').decode(data).match(/#ENCODING=(.+)/i)?.[1] ?? encoding;

    const content = new TextDecoder(encoding || 'utf-8').decode(data);

    const metadata = new Map<string, string>();
    const lines = content.split(/\r\n|\r|\n/g).map(e => e.trim());
    const blocks = [] as Array<{
        start: number;
        textStart: number;
        marker?: string;
        lines?: any[];
    }>;
    let numLanguages = 0;
    let langIndex = 0;

    for (let i = 0; i < lines.length; i++) {
        let line = lines[i];
        if (line === '---' || line === '--' || (allLanguages && line === '--A')) {
            blocks.push({
                start: i + 1,
                textStart: i + 1,
                lines: []
            });
            continue;
        }
        if (line === '--A') continue;

        // Parse Meta Block
        if (!blocks.length) {
            if (line.startsWith('#')) {
                const [key, ...rest] = line.substring(1).split('=');
                metadata.set(key, rest.join('='));
            }
            continue;
        }

        const block = blocks[blocks.length - 1];
        if (!block) continue;

        // Parse Verse Marker
        if (i === block.start) {
            if (line.startsWith('$$M=')) {
                block.marker = line.substring(4);
                continue;
            }
            const verseMarkerMatch = VERSE_MARKERS.find(e => line.startsWith(e));
            if (verseMarkerMatch) {
                const remainder = line.substring(verseMarkerMatch.length).trim();
                if (!remainder || !isNaN(parseInt(remainder))) {
                    block.marker = line || '';
                    block.textStart++;
                    continue;
                }
            }
        }
        numLanguages ||= parseInt(metadata.get('LangCount') || '1');

        if (line.startsWith('#H ')) continue;
        let currentLang = langIndex % numLanguages;

        const langChange = line.match(/^##(\d)/);
        if (langChange) {
            currentLang = parseInt(langChange[1]) - 1;
            line = line.replace(/^##(\d)/, '').trim();
        } else langIndex++;

        // Handle formatting
        line = line
            .replace(/^#C /, '')
            .replace(/<\/?(align|valign|font|h|s|linespacing|posx|incposy|img):?(.+)?\/?>/g, '')
            .replace(/<(c|bgcolor):#(.+?)>/g, '<span class="sb-color" style="--sb-$1: #$2">')
            .replace(/<(c|bgcolor):\$(.{2})(.{2})(.{2})>/g, '<span class="sb-color" style="--sb-$1: #$4$3$2">')
            .replace(/<(c|bgcolor):(.+?)>/g, '<span class="sb-color" style="--sb-$1: var(--sb-$1--$2)">')
            .replace(/<frame(?::(.+?))?(?::(.+?))?>/g, '<span class="sb-frame" style="--sb-frame-w: $1; --sb-frame-p: $2;">')
            .replace(/<wordwrap>/g, '<span class="sb-wordwrap">')
            .replace(/<(\/)?super>/g, '<$1sup>')
            .replace(/<->/g, '')
            .replace(/<\/(c|bgcolor|wordwrap|frame)(:.+?)?>/g, '</span>');

        block.lines?.push({
            text: line,
            language: currentLang
        });
    }

    let slides = blocks;
    if (metadata.has('VerseOrder')) {
        slides = [];
        const verseOrder = metadata.get('VerseOrder')!.split(',');
        for (const verse of verseOrder) {
            const block = blocks.find(b => b.marker === verse);
            if (block) slides.push(block);
        }
    }

    return slides;
}
