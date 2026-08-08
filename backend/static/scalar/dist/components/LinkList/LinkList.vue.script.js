import { createElementBlock, defineComponent, normalizeClass, onMounted, onUnmounted, openBlock, ref, renderSlot } from "vue";
//#region src/components/LinkList/LinkList.vue?vue&type=script&setup=true&lang.ts
var LinkList_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "LinkList",
	setup(__props) {
		const containerRef = ref();
		/** Whether the container needs to scroll */
		const needsScroll = ref(false);
		/** Check if the container can scroll in either direction */
		const checkScrollability = () => {
			if (!containerRef.value) return;
			const { scrollWidth, clientWidth } = containerRef.value;
			needsScroll.value = scrollWidth > clientWidth;
		};
		/** MutationObserver to watch for changes in child elements */
		let mutationObserver = null;
		/**
		* We use the mutation observer to watch for changes and check if we need to scroll,
		* if we do need to scroll we apply the icons-only class to the container
		*/
		onMounted(() => {
			checkScrollability();
			window.addEventListener("resize", checkScrollability);
			if (containerRef.value) {
				mutationObserver = new MutationObserver(() => {
					checkScrollability();
				});
				mutationObserver.observe(containerRef.value, {
					childList: true,
					subtree: true
				});
			}
		});
		onUnmounted(() => {
			window.removeEventListener("resize", checkScrollability);
			if (mutationObserver) {
				mutationObserver.disconnect();
				mutationObserver = null;
			}
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				ref_key: "containerRef",
				ref: containerRef,
				class: normalizeClass(["custom-scroll narrow:mb-3 mb-1.5 flex h-auto min-h-8 max-w-full items-center gap-2 overflow-x-auto text-base whitespace-nowrap", { "icons-only": needsScroll.value }])
			}, [renderSlot(_ctx.$slots, "default", {}, void 0, true)], 2);
		};
	}
});
//#endregion
export { LinkList_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=LinkList.vue.script.js.map