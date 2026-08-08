import { useLocalization } from "../../localization/use-localization.js";
import ApiReferenceToolbarPopover_default from "./ApiReferenceToolbarPopover.vue.js";
import { createBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { ScalarIconCopy, ScalarIconInfo } from "@scalar/icons";
import { ScalarIconButton } from "@scalar/components/icon-button";
//#region src/features/developer-tools/components/ApiReferenceToolbarTitle.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	class: "text-c-2 hover:text-c-1 hover:bg-b-2 ml-auto flex items-center gap-1 rounded px-2 py-2.25 text-base leading-none",
	type: "button"
};
var _hoisted_2 = { class: "-m-2 flex flex-col gap-2 leading-relaxed" };
var _hoisted_3 = { class: "bg-b-2 inline-flex items-center gap-0.5 rounded border px-1 py-0.5 text-sm" };
var CONFIG_SETTING = "showDeveloperTools: \"never\"";
var ApiReferenceToolbarTitle_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarTitle",
	setup(__props) {
		const { copyToClipboard } = useClipboard();
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(ApiReferenceToolbarPopover_default, {
				class: "w-120",
				placement: "bottom-start"
			}, {
				button: withCtx(() => [createElementVNode("button", _hoisted_1, [createVNode(unref(ScalarIconInfo)), createTextVNode(" " + toDisplayString(unref(translate)("developerTools.title")), 1)])]),
				info: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.localhostOnly")), 1)]),
				default: withCtx(() => [createElementVNode("div", _hoisted_2, [createElementVNode("div", null, toDisplayString(unref(translate)("developerTools.intro")), 1), createElementVNode("div", null, [
					createTextVNode(toDisplayString(unref(translate)("developerTools.disableToolbarBefore")) + " ", 1),
					createElementVNode("div", _hoisted_3, [createElementVNode("code", { class: "font-code" }, toDisplayString(CONFIG_SETTING)), createVNode(unref(ScalarIconButton), {
						class: "-m-1 p-1.25",
						icon: unref(ScalarIconCopy),
						label: unref(translate)("actions.copyToClipboard"),
						size: "sm",
						onClick: _cache[0] || (_cache[0] = ($event) => unref(copyToClipboard)(CONFIG_SETTING))
					}, null, 8, ["icon", "label"])]),
					createTextVNode(" " + toDisplayString(unref(translate)("developerTools.disableToolbarAfter")), 1)
				])])]),
				_: 1
			});
		};
	}
});
//#endregion
export { ApiReferenceToolbarTitle_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarTitle.vue.script.js.map