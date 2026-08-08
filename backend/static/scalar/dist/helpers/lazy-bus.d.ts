/** Tracks when the initial load is complete. */
export declare const firstLazyLoadComplete: import("vue").Ref<boolean, boolean>;
/**
 * The id of the element we are currently scrolling to (the anchor target).
 *
 * Schema properties live inside collapsible disclosures that are not part of the
 * navigation tree, so the lazy bus cannot expand them the way it expands sidebar
 * parents. Instead we publish the active target here and let each collapsible
 * schema disclosure open itself when its breadcrumb is on the path to the target.
 * This is what makes deep links to hidden (collapsed) schema properties work.
 */
export declare const scrollTargetId: import("vue").Ref<string, string>;
export declare const getLazyPlaceholderHeight: (id: string) => number | undefined;
export declare const setLazyPlaceholderHeight: (id: string, height: number) => void;
type UnblockFn = () => void;
/**
 * Blocks intersection until the returned unblock callback is run.
 * Prevents scroll jump while we render new lazy content.
 */
export declare const blockIntersection: () => UnblockFn;
/** If there are any pending blocking operations we disable intersection */
export declare const intersectionEnabled: import("vue").ComputedRef<boolean>;
/**
 * Add elements to the priority queue for immediate rendering.
 * We allow adding items already in readyQueue so that callbacks are still triggered,
 * but processQueue will skip actual re-rendering for items already ready.
 */
export declare const addToPriorityQueue: (id: string | undefined) => void;
/**
 * Request an item to be rendered (e.g. when it re-enters the overscan zone).
 */
export declare const requestLazyRender: (id: string | undefined, priority?: boolean) => void;
/**
 * Schedules a single run of the lazy bus so that documents with no Lazy components
 * (e.g. no operations, tags, or models) still get firstLazyLoadComplete set and the
 * full-viewport placeholder can be hidden. Call from content root on mount.
 */
export declare const scheduleInitialLoadComplete: () => void;
/**
 * Tracks the lazy loading state of an element.
 * Use isReady (or expanded) to decide whether to render the slot or show a placeholder.
 * The element is only added to the queue when it enters the viewport overscan (see Lazy.vue).
 */
export declare function useLazyBus(id: string): {
    isReady: import("vue").ComputedRef<boolean>;
};
/**
 * Scroll to a possibly lazy-loaded element. Expands parents and adds target (and
 * parents) to the priority queue, then scrolls after Vue has flushed.
 */
export declare const scrollToLazy: (id: string, setExpanded: (id: string, value: boolean) => void, getEntryById: (id: string) => {
    id: string;
    parent?: {
        id: string;
    };
    children?: {
        id: string;
    }[];
} | undefined) => void;
export {};
//# sourceMappingURL=lazy-bus.d.ts.map