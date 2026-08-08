import { getSchemaParamsFromId } from "./id-routing.js";
import { computed, nextTick, onBeforeUnmount, reactive, ref } from "vue";
import { watchDebounced } from "@vueuse/core";
import { nanoid } from "nanoid";
//#region src/helpers/lazy-bus.ts
/**
* List of items that are in the priority queue and will be rendered first (e.g. scroll target).
*/
var priorityQueue = reactive(/* @__PURE__ */ new Set());
/** List of items that are pending to be loaded (in viewport overscan). */
var pendingQueue = reactive(/* @__PURE__ */ new Set());
/** List of items that are already loaded and stay mounted (no eviction). */
var readyQueue = reactive(/* @__PURE__ */ new Set());
/**
* Flag to indicate if the lazy bus is currently running
* Blocks ID changes while running
*/
var isRunning = ref(false);
/** How long tryScroll keeps retrying to find the element (ms). */
var SCROLL_RETRY_MS = 3e3;
/** Tracks when the initial load is complete. */
var firstLazyLoadComplete = ref(false);
/**
* The id of the element we are currently scrolling to (the anchor target).
*
* Schema properties live inside collapsible disclosures that are not part of the
* navigation tree, so the lazy bus cannot expand them the way it expands sidebar
* parents. Instead we publish the active target here and let each collapsible
* schema disclosure open itself when its breadcrumb is on the path to the target.
* This is what makes deep links to hidden (collapsed) schema properties work.
*/
var scrollTargetId = ref("");
/**
* Clears the scroll target once we are done with it, so a stale target cannot
* re-open a disclosure the user later collapsed (or that remounts). Guarded by
* the id so a newer navigation that started during the scroll retry is kept.
*/
var clearScrollTarget = (id) => {
	if (scrollTargetId.value === id) scrollTargetId.value = "";
};
/** List of unique identifiers that are blocking intersection */
var intersectionBlockers = reactive(/* @__PURE__ */ new Set());
var onRenderComplete = /* @__PURE__ */ new Set();
/** Cached content heights so placeholders can match when not rendered. */
var lazyPlaceholderHeights = reactive(/* @__PURE__ */ new Map());
var getLazyPlaceholderHeight = (id) => lazyPlaceholderHeights.get(id);
var setLazyPlaceholderHeight = (id, height) => {
	if (!Number.isFinite(height) || height <= 0) return;
	lazyPlaceholderHeights.set(id, Math.round(height));
};
/** Adds a one time callback to be executed when the lazy bus has finished loading */
var addLazyCompleteCallback = (callback) => {
	if (callback) onRenderComplete.add(callback);
};
/**
* Blocks intersection until the returned unblock callback is run.
* Prevents scroll jump while we render new lazy content.
*/
var blockIntersection = () => {
	const blockId = nanoid();
	intersectionBlockers.add(blockId);
	/** Unblock uses a small delay to ensure the scroll is complete before enabling intersection */
	return () => setTimeout(() => intersectionBlockers.delete(blockId), 100);
};
/** If there are any pending blocking operations we disable intersection */
var intersectionEnabled = computed(() => intersectionBlockers.size === 0);
/**
* Processes the full queue: priority first, then pending. Blocks intersection while
* rendering so the viewport does not jump. No eviction — items stay in readyQueue.
*/
var runLazyBus = () => {
	if (typeof window === "undefined") return;
	if (isRunning.value) return;
	isRunning.value = true;
	/**
	* Sets all the pending elements into the ready queue
	* After waiting for Vue to update the DOM we execute the callbacks and unblock intersection
	*/
	const processQueue = async () => {
		const priorityIds = [...priorityQueue];
		const pendingIds = [...pendingQueue];
		if (priorityIds.length === 0 && pendingIds.length === 0) {
			onRenderComplete.forEach((fn) => fn());
			onRenderComplete.clear();
			isRunning.value = false;
			firstLazyLoadComplete.value = true;
			return;
		}
		for (const id of priorityIds) {
			readyQueue.add(id);
			priorityQueue.delete(id);
		}
		for (const id of pendingIds) {
			readyQueue.add(id);
			pendingQueue.delete(id);
		}
		await nextTick();
		onRenderComplete.forEach((fn) => fn());
		onRenderComplete.clear();
		isRunning.value = false;
		firstLazyLoadComplete.value = true;
	};
	if (window.requestIdleCallback) window.requestIdleCallback(processQueue, { timeout: 1500 });
	else nextTick(processQueue);
};
/**
* Run the lazy bus when the queue changes and is not currently running
* Debounce so that multiple changes to the queue are batched together
*
* We must run when the priority queue changes because we rely on finish callbacks
* anytime we request potentially lazy elements. If we don't run when the priority queue changes
* we may not have a finish callback even though the element is set to load.
*/
watchDebounced([
	() => pendingQueue.size,
	() => priorityQueue.size,
	() => isRunning.value
], () => {
	if ((pendingQueue.size > 0 || priorityQueue.size > 0) && !isRunning.value) runLazyBus();
}, {
	debounce: 300,
	maxWait: 1500
});
/**
* We only make elements pending if they are not already in the priority or ready queue
*/
var addToPendingQueue = (id) => {
	if (id && !readyQueue.has(id) && !priorityQueue.has(id)) pendingQueue.add(id);
};
/**
* Add elements to the priority queue for immediate rendering.
* We allow adding items already in readyQueue so that callbacks are still triggered,
* but processQueue will skip actual re-rendering for items already ready.
*/
var addToPriorityQueue = (id) => {
	if (id && !priorityQueue.has(id)) priorityQueue.add(id);
};
/**
* Request an item to be rendered (e.g. when it re-enters the overscan zone).
*/
var requestLazyRender = (id, priority = false) => {
	if (!id || readyQueue.has(id)) return;
	if (priority) addToPriorityQueue(id);
	else addToPendingQueue(id);
	if (!isRunning.value) runLazyBus();
};
/**
* Schedules a single run of the lazy bus so that documents with no Lazy components
* (e.g. no operations, tags, or models) still get firstLazyLoadComplete set and the
* full-viewport placeholder can be hidden. Call from content root on mount.
*/
var scheduleInitialLoadComplete = () => {
	if (typeof window === "undefined") return;
	window.setTimeout(() => runLazyBus(), 400);
};
/** When an element is unmounted we remove it from all queues */
var resetLazyElement = (id) => {
	priorityQueue.delete(id);
	pendingQueue.delete(id);
	readyQueue.delete(id);
	lazyPlaceholderHeights.delete(id);
};
/**
* Tracks the lazy loading state of an element.
* Use isReady (or expanded) to decide whether to render the slot or show a placeholder.
* The element is only added to the queue when it enters the viewport overscan (see Lazy.vue).
*/
function useLazyBus(id) {
	onBeforeUnmount(() => {
		resetLazyElement(id);
	});
	return { isReady: computed(() => typeof window === "undefined" || priorityQueue.has(id) || readyQueue.has(id)) };
}
/**
* Scroll to a possibly lazy-loaded element. Expands parents and adds target (and
* parents) to the priority queue, then scrolls after Vue has flushed.
*/
var scrollToLazy = (id, setExpanded, getEntryById) => {
	const item = getEntryById(id);
	const unfreeze = !readyQueue.has(id) || item?.children?.some((child) => !readyQueue.has(child.id)) ? freeze(id) : void 0;
	addLazyCompleteCallback(unfreeze);
	const unblock = blockIntersection();
	const { rawId } = getSchemaParamsFromId(id);
	scrollTargetId.value = id;
	addToPriorityQueue(id);
	addToPriorityQueue(rawId);
	if (item?.children) item.children.slice(0, 2).forEach((child) => addToPriorityQueue(child.id));
	if (item?.parent) {
		const parent = getEntryById(item.parent.id);
		const elementIdx = parent?.children?.findIndex((child) => child.id === id);
		if (elementIdx !== void 0 && elementIdx >= 0) parent?.children?.slice(elementIdx, elementIdx + 2).forEach((child) => addToPriorityQueue(child.id));
	}
	setExpanded(rawId, true);
	/**
	* Recursively expand the parents and set them as a loading priority
	* This ensures all parents will be immediately loaded and open
	*/
	const addParents = (currentId) => {
		const parent = getEntryById(currentId)?.parent;
		if (parent) {
			addToPriorityQueue(parent.id);
			setExpanded(parent.id, true);
			addParents(parent.id);
		}
	};
	/** Must use the rawId as schema params are not in the navigation tree */
	addParents(rawId);
	nextTick(() => {
		tryScroll(id, Date.now() + SCROLL_RETRY_MS, unblock, unfreeze);
	});
};
/**
* Tiny wrapper around the scrollIntoView API
* Retries up to the stopTime in case the element is not yet rendered
*
* @param id - The id of the element to scroll to
* @param stopTime - The time to stop retrying in unix milliseconds
*/
var tryScroll = (id, stopTime, onComplete, onFailure) => {
	const element = document.getElementById(id);
	if (element) {
		element.scrollIntoView({ block: "start" });
		clearScrollTarget(id);
		onComplete();
	} else if (Date.now() < stopTime) requestAnimationFrame(() => tryScroll(id, stopTime, onComplete, onFailure));
	else {
		clearScrollTarget(id);
		onComplete();
		onFailure?.();
	}
};
var freeze = (id) => {
	let stop = false;
	/**
	* Runs until the stop flag is set
	* Executes the final frame after stop changes to true
	*/
	const runFrame = (stopAfterFrame) => {
		const element = document.getElementById(id);
		if (element) element.scrollIntoView({ block: "start" });
		if (!stopAfterFrame) requestAnimationFrame(() => runFrame(stop));
	};
	runFrame(false);
	return () => {
		stop = true;
	};
};
//#endregion
export { addToPriorityQueue, blockIntersection, firstLazyLoadComplete, getLazyPlaceholderHeight, intersectionEnabled, requestLazyRender, scheduleInitialLoadComplete, scrollTargetId, scrollToLazy, setLazyPlaceholderHeight, useLazyBus };

//# sourceMappingURL=lazy-bus.js.map