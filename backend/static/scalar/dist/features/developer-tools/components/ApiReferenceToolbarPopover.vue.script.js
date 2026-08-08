import { useLocalization } from "../../localization/use-localization.js";
import { createBlock, createElementVNode, createTextVNode, createVNode, defineComponent, guardReactiveProps, normalizeClass, normalizeProps, openBlock, renderSlot, toDisplayString, unref, withCtx } from "vue";
import { ScalarIconCaretDown, ScalarIconInfo } from "@scalar/icons";
import { ScalarFloatingBackdrop } from "@scalar/components/floating";
import { ScalarPopover } from "@scalar/components/popover";
//#region src/features/developer-tools/components/ApiReferenceToolbarPopover.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	class: "text-c-2 hover:text-c-1 hover:bg-b-2 flex items-center gap-1 rounded px-2 py-2.25 text-base leading-none",
	type: "button"
};
var _hoisted_2 = { class: "custom-scroll bg-b-1 flex flex-col gap-7 rounded-lg p-7 pb-6" };
var _hoisted_3 = { class: "text-c-2 flex items-center justify-center gap-1 p-2 text-sm" };
var ApiReferenceToolbarPopover_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarPopover",
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarPopover), {
				class: "max-h-[inherit] max-w-[inherit] p-0 text-base",
				placement: "bottom-end",
				teleport: ""
			}, {
				default: withCtx(({ open }) => [renderSlot(_ctx.$slots, "button", { open }, () => [createElementVNode("button", _hoisted_1, [renderSlot(_ctx.$slots, "label"), createVNode(unref(ScalarIconCaretDown), { class: normalizeClass(["size-3", { "rotate-180": open }]) }, null, 8, ["class"])])])]),
				popover: withCtx((props) => [createElementVNode("div", _hoisted_2, [renderSlot(_ctx.$slots, "default", normalizeProps(guardReactiveProps(props)))]), createElementVNode("div", _hoisted_3, [createVNode(unref(ScalarIconInfo), { class: "size-3.5 shrink-0" }), createElementVNode("div", null, [renderSlot(_ctx.$slots, "info", {}, () => [createTextVNode(toDisplayString(unref(translate)("developerTools.localhostOnly")), 1)])])])]),
				backdrop: withCtx(() => [createVNode(unref(ScalarFloatingBackdrop), { class: "bg-b-2 rounded-lg" })]),
				_: 3
			});
		};
	}
});
//#endregion
export { ApiReferenceToolbarPopover_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarPopover.vue.script.js.map