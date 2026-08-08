import { computed, createElementBlock, defineComponent, mergeProps, openBlock, renderSlot, unref } from "vue";
import { useBindCx } from "@scalar/use-hooks/useBindCx";
//#region src/components/Badge/Badge.vue?vue&type=script&setup=true&lang.ts
var Badge_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	inheritAttrs: false,
	__name: "Badge",
	props: { color: {} },
	setup(__props) {
		const { cx } = useBindCx();
		const badgeStyle = computed(() => __props.color ? {
			backgroundColor: __props.color,
			color: `color-mix(in srgb, ${__props.color}, black 40%)`
		} : void 0);
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", mergeProps(unref(cx)("badge inline-block rounded-2xl border bg-b-2 px-1.5 py-0.5 text-c-2 text-sm"), { style: badgeStyle.value }), [renderSlot(_ctx.$slots, "default", {}, void 0, true)], 16);
		};
	}
});
//#endregion
export { Badge_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Badge.vue.script.js.map