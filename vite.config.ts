import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import { fileURLToPath } from 'node:url';
import { cp } from 'node:fs/promises';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');
    const WR_SERVER = env.WR_SERVER || '127.0.0.1:10024';

    return {
        plugins: [vue(), tailwindcss()],
        build: {
            license: { fileName: 'LICENSE.md' }
        },
        server: {
            proxy: {
                '/osc': {
                    target: `ws://${WR_SERVER}/`,
                    ws: true
                },
                '/sbfiles': {
                    target: `http://${WR_SERVER}/`
                },
                '/thumbnails': {
                    target: `http://${WR_SERVER}/`
                }
            }
        }
    };
});

await cp(fileURLToPath(import.meta.resolve('pdfjs-dist/build/pdf.worker.min.mjs')), './public/pdf.worker.min.js');
