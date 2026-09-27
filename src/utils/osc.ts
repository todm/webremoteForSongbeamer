import { concatBytes } from '.';

export default class OSCMessage {
    private address: string;
    private args: OSCArgument<any>[];

    constructor(address: string, ...args: any[]) {
        this.address = address;
        this.args = this.mapArguments(args);
    }

    getAddress() {
        return this.address;
    }

    getArguments() {
        return [...this.args];
    }

    getValues() {
        return this.args.map(e => e.getValue());
    }

    pack(): ArrayBuffer {
        const parts = [] as ArrayBuffer[];
        parts.push(new OSCString(this.address).pack());
        parts.push(new OSCString(',' + this.args.map(arg => arg.getSymbol()).join('')).pack());
        parts.push(...this.args.map(arg => arg.pack()));
        return concatBytes(...parts.map(e => new Uint8Array(e))).buffer;
    }

    unpack(data: ArrayBuffer): number {
        let offset = 0;

        const addressType = new OSCString('');
        offset += addressType.unpack(data.slice(offset));
        this.address = addressType.getValue();

        const typeTagType = new OSCString('');
        offset += typeTagType.unpack(data.slice(offset));
        const typeTags = typeTagType.getValue().substring(1); // Remove leading ','

        this.args = typeTags.split('').map(t => {
            let arg: OSCArgument<any>;
            switch (t) {
                case 's':
                    arg = new OSCString('');
                    break;
                case 'i':
                    arg = new OSCInt(0);
                    break;
                case 'f':
                    arg = new OSCFloat(0);
                    break;
                case 'd':
                    arg = new OSCDouble(0);
                    break;
                case 'b':
                    arg = new OSCBlob(new Uint8Array());
                    break;
                case 'r':
                    arg = new OSCColor(0);
                    break;
                default:
                    throw new Error('unsupported type: ' + t);
            }
            offset += arg.unpack(data.slice(offset));
            return arg;
        });

        return offset;
    }

    private mapArguments(args: any[]): OSCArgument<any>[] {
        return args.map(arg => {
            if (arg instanceof OSCArgument) return arg;

            if (typeof arg === 'string') {
                return new OSCString(arg);
            }
            if (typeof arg === 'number') {
                if (Number.isInteger(arg)) return new OSCInt(arg);
                else return new OSCFloat(arg);
            }
            if (arg instanceof Uint8Array) {
                return new OSCBlob(arg);
            }
            throw new Error('unknown argument type');
        });
    }
}

abstract class OSCArgument<T> {
    protected value: T;
    protected readonly symbol = '';
    constructor(value: T) {
        this.value = value;
    }

    getValue(): T {
        return this.value;
    }

    abstract getSymbol(): string;
    abstract pack(): ArrayBuffer;
    abstract unpack(data: ArrayBuffer): number;
}

export class OSCString extends OSCArgument<string> {
    getSymbol(): string {
        return 's';
    }
    pack(): ArrayBuffer {
        const encoded = new TextEncoder().encode(this.value + '\0');
        const data = new Uint8Array((encoded.length + 3) & ~3);
        data.set(encoded);
        return data.buffer;
    }
    unpack(data: ArrayBuffer): number {
        const buf = new Uint8Array(data);
        const end = buf.indexOf(0);
        if (end === -1) throw new Error('osc string not terminated');
        this.value = new TextDecoder().decode(buf.subarray(0, end));
        return (end + 4) & ~3;
    }
}

export class OSCInt extends OSCArgument<number> {
    getSymbol(): string {
        return 'i';
    }
    pack(): ArrayBuffer {
        const buffer = new ArrayBuffer(4);
        const view = new DataView(buffer);
        view.setInt32(0, this.value, false);
        return buffer;
    }
    unpack(data: ArrayBuffer): number {
        const view = new DataView(data);
        this.value = view.getInt32(0, false);
        return 4;
    }
}

export class OSCFloat extends OSCArgument<number> {
    getSymbol(): string {
        return 'f';
    }
    pack(): ArrayBuffer {
        const buffer = new ArrayBuffer(4);
        const view = new DataView(buffer);
        view.setFloat32(0, this.value, false);
        return buffer;
    }
    unpack(data: ArrayBuffer): number {
        const view = new DataView(data);
        this.value = view.getFloat32(0, false);
        return 4;
    }
}

export class OSCDouble extends OSCArgument<number> {
    getSymbol(): string {
        return 'd';
    }
    pack(): ArrayBuffer {
        const buffer = new ArrayBuffer(8);
        const view = new DataView(buffer);
        view.setFloat64(0, this.value, false);
        return buffer;
    }
    unpack(data: ArrayBuffer) {
        const view = new DataView(data);
        this.value = view.getFloat64(0, false);
        return 8;
    }
}

export class OSCBlob extends OSCArgument<Uint8Array> {
    getSymbol(): string {
        return 'b';
    }
    pack(): ArrayBuffer {
        const data = new Uint8Array(4 + ((this.value.length + 3) & ~3));
        const view = new DataView(data.buffer);
        view.setInt32(0, this.value.length, false);
        data.set(this.value, 4);
        return data.buffer;
    }
    unpack(data: ArrayBuffer) {
        const view = new DataView(data);
        const length = view.getInt32(0, false);
        this.value = new Uint8Array(data, 4, length);
        return 4 + ((length + 3) & ~3);
    }
}

export class OSCColor extends OSCInt {
    getSymbol(): string {
        return 'r';
    }
}
