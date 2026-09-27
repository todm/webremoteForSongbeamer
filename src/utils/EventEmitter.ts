/**
 * A simple EventEmitter class to handle event-driven programming.
 * It allows registering event listeners, emitting events, and removing listeners.
 */
export default class EventEmitter<T extends Record<string, any[]>> {
    private listeners: { [K in keyof T]?: Array<(...args: T[K]) => void> } = {};

    /**
     * Registers an event listener for the specified event.
     * @param event The name of the event to listen for.
     * @param listener The callback function to invoke when the event is emitted.
     * @returns A function to unregister the listener.
     */
    on<K extends keyof T>(event: K, listener: (...args: T[K]) => void) {
        this.listeners[event] ??= [];
        this.listeners[event]!.push(listener);
        return () => this.off(event, listener);
    }

    /**
     * Registers a one-time event listener for the specified event.
     * @param event The name of the event to listen for.
     * @param listener The callback function to invoke when the event is emitted.
     * @returns A function to unregister the listener.
     */
    once<K extends keyof T>(event: K, listener: (...args: T[K]) => void) {
        const onceListener = (...args: T[K]) => {
            listener(...args);
            this.off(event, onceListener);
        };
        return this.on(event, onceListener);
    }

    /**
     * Unregisters an event listener for the specified event.
     * @param event The name of the event.
     * @param listener The callback function to remove.
     * @return The EventEmitter instance for chaining.
     */
    off<K extends keyof T>(event: K, listener: (...args: T[K]) => void) {
        this.listeners[event] = this.listeners[event]?.filter(l => l !== listener);
        return this;
    }

    /**
     * Emits an event, invoking all registered listeners with the provided arguments.
     * @param event The name of the event to emit.
     * @param args The arguments to pass to the event listeners.
     * @returns The EventEmitter instance for chaining.
     */
    protected emit<K extends keyof T>(event: K, ...args: T[K]) {
        this.listeners[event]?.forEach(listener => listener(...args));
        return this;
    }

    /**
     * Removes all listeners for the specified event, or all listeners if no event is specified.
     * @param event The name of the event to remove listeners for (optional).
     * @return The EventEmitter instance for chaining.
     */
    removeAllListeners(event?: keyof T) {
        if (event) {
            this.listeners[event] = [];
        } else {
            this.listeners = {};
        }
        return this;
    }
}