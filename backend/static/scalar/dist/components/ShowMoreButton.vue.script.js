import { useLocalization } from "../features/localization/use-localization.js";
import { createElementBlock, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconCaretDown } from "@scalar/icons";
//#region src/components/ShowMoreButton.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	class: "show-more",
	type: "button"
};
var ShowMoreButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ShowMoreButton",
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("button", _hoisted_1, [createTextVNode(toDisplayString(unref(translate)("actions.showMore")) + " ", 1), createVNode(unref(ScalarIconCaretDown), {
				class: "text-c-2 mt-0.25 size-3",
				weight: "bold"
			})]);
		};
	}
});
//#endregion
export { ShowMoreButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ShowMoreButton.vue.script.js.map