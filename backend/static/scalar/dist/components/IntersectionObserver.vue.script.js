import { createBlock, defineComponent, onMounted, openBlock, ref, renderSlot, resolveDynamicComponent, withCtx } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
//#region src/components/IntersectionObserver.vue?vue&type=script&setup=true&lang.ts
var IntersectionObserver_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "IntersectionObserver",
	props: {
		id: {},
		is: {}
	},
	emits: ["intersecting"],
	setup(__props, { emit: __emit }) {
		const props = __props;
		const emit = __emit;
		const intersectionObserverRef = ref();
		const calculateRootMargin = (element) => {
			const height = element.offsetHeight;
			return `${height / 2}px 0px ${height / 2}px 0px`;
		};
		const calculateThreshold = (element) => {
			return element.offsetHeight < window.innerHeight ? .8 : .5;
		};
		onMounted(() => {
			if (intersectionObserverRef.value) {
				const options = {
					rootMargin: calculateRootMargin(intersectionObserverRef.value),
					threshold: calculateThreshold(intersectionObserverRef.value)
				};
				useIntersectionObserver(intersectionObserverRef, ([entry]) => {
					if (entry?.isIntersecting && props.id) emit("intersecting", props.id);
				}, options);
			}
		});
		return (_ctx, _cache) => {
			return openBlock(), createBlock(resolveDynamicComponent(__props.is ?? "div"), {
				id: __props.id,
				ref_key: "intersectionObserverRef",
				ref: intersectionObserverRef
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default")]),
				_: 3
			}, 8, ["id"]);
		};
	}
});
//#endregion
export { IntersectionObserver_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=IntersectionObserver.vue.script.js.map