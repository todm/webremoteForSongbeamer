import { EventEmitter, getValueAtPath } from '../utils';
import OSCMessage from '../utils/osc';

export default class OscClient extends EventEmitter<{ open: []; error: []; close: []; message: [path: string, args: any[]] }> {
    private ws?: WebSocket;
    private url: string;
    private isOpen = false;
    private reconnectionTimeout?: ReturnType<typeof setTimeout>;
    private enableLog = false;

    constructor(url: string) {
        super();
        this.url = url;

        this.enableLog = !!localStorage.getItem('osc-log');
    }

    public connect() {
        this.ws?.close();
        this.ws = new WebSocket(this.url);
        this.isOpen = true;
        this.ws.addEventListener('open', () => this.emit('open'));
        this.ws.addEventListener('error', () => this.emit('error'));
        this.ws.addEventListener('close', this.onClose.bind(this));
        this.ws.addEventListener('message', this.onMessage.bind(this));
        return this;
    }

    public disconnect() {
        this.isOpen = false;
        this.ws?.close();
        this.removeAllListeners();
        clearTimeout(this.reconnectionTimeout);
        this.ws = undefined;
        return this;
    }

    public status() {
        return this.ws?.readyState;
    }

    private onClose() {
        if (this.isOpen) this.reconnectionTimeout = setTimeout(() => this.connect(), 1000);
        else this.emit('close');
    }

    private async onMessage(e: MessageEvent) {
        const data: ArrayBuffer = await e.data.arrayBuffer();
        const message = new OSCMessage('');
        message.unpack(data);
        this.enableLog && console.log('🟡 OSC', message.getAddress(), ...message.getValues());
        this.emit('message', message.getAddress(), message.getValues());
    }

    public async send(address: string, ...args: any[]) {
        if (!this.isOpen) return this;
        if (this.ws?.readyState !== WebSocket.OPEN) {
            setTimeout(() => this.send(address, ...args), 100);
            return this;
        }
        const message = new OSCMessage(address, ...args);
        const data = message.pack();
        this.ws?.send(data);
        this.enableLog && console.log('🟢 OSC', address, ...args);
        return this;
    }

    public async request(address: string, ...args: any[]) {
        return new Promise<{ path: string; args: any[] }>((resolve, reject) => {
            const c = new OscClient(this.url);

            const timeout = setTimeout(() => {
                reject(new Error('Request timed out'));
                c.disconnect();
            }, 10000);

            c.on('message', (path, args) => {
                resolve({ path, args });
                c.disconnect();
                clearTimeout(timeout);
            });

            c.connect();
            c.send(address, ...args);
        });
    }
}

export class XOscClient extends OscClient {
    private heartbeatInterval?: ReturnType<typeof setInterval>;

    constructor(url: string) {
        super(url);

        this.on('open', this.subscribe.bind(this));
        this.on('close', this.unsubscribe.bind(this));
    }

    private subscribe() {
        this.send('/xremote');
        this.heartbeatInterval = setInterval(() => {
            this.send('/xremote');
        }, 8000);
    }

    private unsubscribe() {
        clearInterval(this.heartbeatInterval);
    }

    public sync(obj: Record<string, any>, readOnly = false, path: string = '/') {
        const val = getValueAtPath(obj, path);
        if (typeof val === 'object' && val !== null) {
            for (const key in val) {
                if (key.startsWith('_')) continue;
                this.sync(obj, readOnly, `${path}${path.at(-1) === '/' ? '' : '/'}${key}`);
            }
            return this;
        }

        if (readOnly) this.send(path);
        else this.send(path, val);
        return this;
    }
}
