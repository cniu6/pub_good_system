import { useLocalization } from "../../../features/localization/use-localization.js";
import { formatValue } from "./helpers/format-value.js";
import { createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { ScalarIcon } from "@scalar/components/icon";
//#region src/components/Content/Schema/SchemaPropertyDefault.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "property-default"
};
var _hoisted_2 = {
	class: "property-default-label",
	type: "button"
};
var _hoisted_3 = { class: "property-default-value-list" };
var SchemaPropertyDefault_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaPropertyDefault",
	props: { value: {} },
	setup(__props) {
		const { copyToClipboard } = useClipboard();
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return __props.value !== void 0 ? (openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("button", _hoisted_2, [createElementVNode("span", null, toDisplayString(unref(translate)("schema.default")), 1)]), createElementVNode("div", _hoisted_3, [createElementVNode("button", {
				class: "property-default-value group",
				type: "button",
				onClick: _cache[0] || (_cache[0] = ($event) => unref(copyToClipboard)(unref(formatValue)(__props.value)))
			}, [createElementVNode("span", null, toDisplayString(unref(formatValue)(__props.value)), 1), createVNode(unref(ScalarIcon), {
				class: "group-hover:text-c-1 text-c-3 ml-auto min-h-3 min-w-3",
				icon: "Clipboard",
				size: "xs"
			})])])])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SchemaPropertyDefault_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaPropertyDefault.vue.script.js.map