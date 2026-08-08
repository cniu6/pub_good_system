import { useLocalization } from "../../features/localization/use-localization.js";
import ScreenReader_default from "../ScreenReader.vue.js";
import { createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, guardReactiveProps, normalizeProps, openBlock, renderSlot, toDisplayString, unref, useId, withCtx, withModifiers } from "vue";
import { ScalarIconHash } from "@scalar/icons";
import { ScalarButton } from "@scalar/components/button";
import { useBindCx } from "@scalar/use-hooks/useBindCx";
//#region src/components/Anchor/Anchor.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var _hoisted_2 = { class: "relative" };
var Anchor_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Anchor",
	emits: ["copyAnchorUrl"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		const labelId = useId();
		const { translate } = useLocalization();
		const { cx } = useBindCx();
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("span", normalizeProps(guardReactiveProps(unref(cx)("group/heading wrap-break-word relative"))), [createElementVNode("span", {
				id: unref(labelId),
				class: "contents"
			}, [renderSlot(_ctx.$slots, "default")], 8, _hoisted_1), createElementVNode("span", _hoisted_2, [_cache[1] || (_cache[1] = createElementVNode("span", null, "​", -1)), createVNode(unref(ScalarButton), {
				"aria-describedby": unref(labelId),
				class: "absolute top-1/2 left-0 inline-block h-fit -translate-y-1/2 px-1.5 py-1 opacity-0 group-hover/heading:opacity-100 group-has-focus-visible/heading:opacity-100",
				variant: "ghost",
				onClick: _cache[0] || (_cache[0] = withModifiers(() => emit("copyAnchorUrl"), ["stop"]))
			}, {
				default: withCtx(() => [createVNode(unref(ScalarIconHash), {
					"aria-hidden": "true",
					class: "size-4.5"
				}), createVNode(ScreenReader_default, null, {
					default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("actions.copyLink")), 1)]),
					_: 1
				})]),
				_: 1
			}, 8, ["aria-describedby"])])], 16);
		};
	}
});
//#endregion
export { Anchor_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Anchor.vue.script.js.map