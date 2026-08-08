import { getLazyPlaceholderHeight, requestLazyRender, setLazyPlaceholderHeight, useLazyBus } from "../../helpers/lazy-bus.js";
import { computed, createCommentVNode, createElementBlock, defineComponent, nextTick, normalizeStyle, onBeforeUnmount, onMounted, openBlock, ref, renderSlot, watch } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
//#region src/components/Lazy/Lazy.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id", "data-placeholder"];
var PLACEHOLDER_HEIGHT_PX = 760;
/** Overscan: render items within this many pixels above and below the viewport. */
var VIEWPORT_OVERSCAN_PX = 1200;
var Lazy_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Lazy",
	props: {
		id: {},
		expanded: {
			type: Boolean,
			default: false
		}
	},
	setup(__props) {
		/**
		* Lazily renders content when the element is within the viewport overscan.
		* Uses a fixed-height placeholder so layout is stable while we block during render.
		*
		* When server-side rendering, content renders immediately.
		*
		* @link https://medium.com/js-dojo/lazy-rendering-in-vue-to-improve-performance-dcccd445d5f
		*/
		/** Fixed height for all placeholders so we do not measure or jump. */
		const VIEWPORT_ROOT_MARGIN = `${VIEWPORT_OVERSCAN_PX}px 0px`;
		const { isReady } = useLazyBus(__props.id);
		const lazyContainerRef = ref(null);
		const placeholderHeight = ref(getLazyPlaceholderHeight(__props.id) ?? PLACEHOLDER_HEIGHT_PX);
		let contentResizeObserver = null;
		/** Once ready we always show (no eviction). Otherwise show when expanded (e.g. so child Lazy placeholders mount). */
		const shouldRender = computed(() => isReady.value || __props.expanded);
		onMounted(() => {
			if (typeof window === "undefined") return;
			if (!("IntersectionObserver" in window)) {
				requestLazyRender(__props.id, true);
				return;
			}
			useIntersectionObserver(lazyContainerRef, ([entry]) => {
				if (entry?.isIntersecting && !isReady.value) requestLazyRender(__props.id, true);
			}, { rootMargin: VIEWPORT_ROOT_MARGIN });
		});
		/**
		* Capture content height right before we switch to placeholder (pre-flush so content
		* is still in the DOM). Ensures we never measure the container or leave the cache stale.
		*/
		watch(() => shouldRender.value, (rendered, wasRendered) => {
			if (wasRendered && !rendered && lazyContainerRef.value) {
				const h = lazyContainerRef.value.offsetHeight;
				if (Number.isFinite(h) && h > 0) {
					placeholderHeight.value = h;
					setLazyPlaceholderHeight(__props.id, h);
				}
			}
		}, { flush: "pre" });
		/** When content is visible, set up ResizeObserver and measure. */
		watch(() => shouldRender.value, (rendered) => {
			if (!rendered) {
				contentResizeObserver?.disconnect();
				contentResizeObserver = null;
				return;
			}
			nextTick(() => {
				if (!lazyContainerRef.value || typeof ResizeObserver === "undefined") return;
				if (!contentResizeObserver) contentResizeObserver = new ResizeObserver(() => {
					if (!lazyContainerRef.value) return;
					const h = lazyContainerRef.value.offsetHeight;
					if (Number.isFinite(h) && h > 0) {
						placeholderHeight.value = h;
						setLazyPlaceholderHeight(__props.id, h);
					}
				});
				contentResizeObserver.observe(lazyContainerRef.value);
				const h = lazyContainerRef.value.offsetHeight;
				if (Number.isFinite(h) && h > 0) {
					placeholderHeight.value = h;
					setLazyPlaceholderHeight(__props.id, h);
				}
			});
		}, { immediate: true });
		onBeforeUnmount(() => {
			contentResizeObserver?.disconnect();
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				id: !shouldRender.value ? __props.id : void 0,
				ref_key: "lazyContainerRef",
				ref: lazyContainerRef,
				"data-placeholder": !shouldRender.value,
				"data-testid": "lazy-container",
				style: normalizeStyle({ height: shouldRender.value ? void 0 : `${placeholderHeight.value}px` })
			}, [shouldRender.value ? renderSlot(_ctx.$slots, "default", {}, void 0, void 0, 0) : createCommentVNode("", true)], 12, _hoisted_1);
		};
	}
});
//#endregion
export { Lazy_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Lazy.vue.script.js.map