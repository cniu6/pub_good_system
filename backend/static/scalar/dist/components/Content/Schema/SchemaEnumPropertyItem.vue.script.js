import { createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, unref } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
import { ScalarWrappingText } from "@scalar/components/wrapping-text";
//#region src/components/Content/Schema/SchemaEnumPropertyItem.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "property-enum-value" };
var _hoisted_2 = { class: "property-enum-value-content" };
var _hoisted_3 = { class: "property-enum-value-label" };
var _hoisted_4 = {
	key: 0,
	class: "property-enum-value-description"
};
var SchemaEnumPropertyItem_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaEnumPropertyItem",
	props: {
		label: {},
		description: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("li", _hoisted_1, [createElementVNode("div", _hoisted_2, [createElementVNode("span", _hoisted_3, [createVNode(unref(ScalarWrappingText), {
				text: __props.label,
				preset: "property"
			}, null, 8, ["text"])]), __props.description ? (openBlock(), createElementBlock("span", _hoisted_4, [createVNode(unref(ScalarMarkdown), { value: __props.description }, null, 8, ["value"])])) : createCommentVNode("", true)])]);
		};
	}
});
//#endregion
export { SchemaEnumPropertyItem_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaEnumPropertyItem.vue.script.js.map