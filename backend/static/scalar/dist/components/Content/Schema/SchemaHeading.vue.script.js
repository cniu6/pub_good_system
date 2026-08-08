import { useLocalization } from "../../../features/localization/use-localization.js";
import { isTypeObject } from "./helpers/is-type-object.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
import { ScalarWrappingText } from "@scalar/components/wrapping-text";
//#region src/components/Content/Schema/SchemaHeading.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "schema-type"
};
var _hoisted_2 = ["title"];
var SchemaHeading_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaHeading",
	props: {
		value: {},
		name: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/** Generate a failsafe type from the properties when we don't have one */
		const failsafeType = computed(() => {
			if ("type" in __props.value) return __props.value.type;
			if (__props.value.enum) return "enum";
			if (isArraySchema(__props.value) && __props.value.items) return "array";
			if (isTypeObject(__props.value) && (__props.value.properties || __props.value.additionalProperties)) return "object";
			return "unknown";
		});
		return (_ctx, _cache) => {
			return typeof __props.value === "object" ? (openBlock(), createElementBlock("span", _hoisted_1, [createElementVNode("span", {
				class: "schema-type-icon",
				title: "type" in __props.value && typeof __props.value.type === "string" ? __props.value.type : "type" in __props.value && Array.isArray(__props.value.type) ? __props.value.type.join(" | ") : unref(translate)("schema.unknownType")
			}, [
				unref(isTypeObject)(__props.value) ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createTextVNode(" {} ")], 64)) : createCommentVNode("", true),
				unref(isArraySchema)(__props.value) ? (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(" [] ")], 64)) : createCommentVNode("", true),
				__props.value.enum ? (openBlock(), createElementBlock(Fragment, { key: 2 }, [createTextVNode(" enum ")], 64)) : createCommentVNode("", true)
			], 8, _hoisted_2), __props.name ? (openBlock(), createBlock(unref(ScalarWrappingText), {
				key: 0,
				preset: "property",
				text: __props.name
			}, null, 8, ["text"])) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(failsafeType.value), 1)], 64))])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SchemaHeading_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaHeading.vue.script.js.map