import { onMounted } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
//#region src/hooks/use-intersection.ts
/**
* Shrinks the observation root to a thin horizontal strip at the vertical middle of
* the viewport (~2% of viewport height). Emits only while the target overlaps that line.
*/
var VIEWPORT_VERTICAL_CENTER_ROOT_MARGIN = "-49% 0px -49% 0px";
var useIntersection = (el, onIntersect, options) => {
	onMounted(() => {
		const observerOptions = {
			rootMargin: options?.immediate ? "0px 0px 0px 0px" : VIEWPORT_VERTICAL_CENTER_ROOT_MARGIN,
			threshold: 0
		};
		if (el.value) useIntersectionObserver(el, ([entry]) => {
			if (entry?.isIntersecting) onIntersect();
			else options?.onExit?.();
		}, observerOptions);
	});
};
//#endregion
export { useIntersection };

//# sourceMappingURL=use-intersection.js.map