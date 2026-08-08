import { useLocalization } from "../../../features/localization/use-localization.js";
import LinkButton_default from "./LinkButton.vue.js";
import { formatExample } from "./helpers/format-example.js";
import { Fragment, computed, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { isObject } from "@scalar/helpers/object/is-object";
import { ScalarIcon } from "@scalar/components/icon";
//#region src/components/Content/Schema/SchemaPropertyExamples.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "property-example"
};
var _hoisted_2 = { class: "property-example-value-list" };
var _hoisted_3 = {
	key: 1,
	class: "property-example"
};
var _hoisted_4 = { class: "property-example-value-list" };
var _hoisted_5 = ["onClick"];
var SchemaPropertyExamples_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaPropertyExamples",
	props: {
		examples: {},
		example: {}
	},
	setup(__props) {
		const { copyToClipboard } = useClipboard();
		const { translate } = useLocalization();
		const hasSingleExample = computed(() => __props.example !== void 0);
		const normalizedExamples = computed(() => {
			if (__props.examples && typeof __props.examples === "object") return __props.examples;
			return {};
		});
		const hasMultipleExamples = computed(() => Object.keys(normalizedExamples.value).length > 0);
		const multipleExamplesLabel = computed(() => Object.keys(normalizedExamples.value).length === 1 ? translate("schema.example") : translate("schema.examples"));
		/**
		* Unwrap an OpenAPI 3 Example Object (`{ value, externalValue, summary, description }`)
		* to the actual sample. Plain values pass through untouched.
		*/
		function unwrapExampleObject(value) {
			if (isObject(value)) {
				if ("value" in value) return value.value;
				if ("externalValue" in value) return value.externalValue;
			}
			return value;
		}
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [hasSingleExample.value ? (openBlock(), createElementBlock("div", _hoisted_1, [createVNode(LinkButton_default, { class: "decoration-dotted" }, {
				default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("schema.example")), 1)]),
				_: 1
			}), createElementVNode("div", _hoisted_2, [createElementVNode("button", {
				class: "property-example-value group",
				type: "button",
				onClick: _cache[0] || (_cache[0] = ($event) => unref(copyToClipboard)(unref(formatExample)(__props.example)))
			}, [createElementVNode("span", null, toDisplayString(unref(formatExample)(__props.example)), 1), createVNode(unref(ScalarIcon), {
				class: "group-hover:text-c-1 text-c-3 ml-auto min-h-3 min-w-3",
				icon: "Clipboard",
				size: "xs"
			})])])])) : createCommentVNode("", true), hasMultipleExamples.value ? (openBlock(), createElementBlock("div", _hoisted_3, [createVNode(LinkButton_default, { class: "decoration-dotted" }, {
				default: withCtx(() => [createTextVNode(toDisplayString(multipleExamplesLabel.value), 1)]),
				_: 1
			}), createElementVNode("div", _hoisted_4, [(openBlock(true), createElementBlock(Fragment, null, renderList(normalizedExamples.value, (ex, key) => {
				return openBlock(), createElementBlock("button", {
					key,
					class: "property-example-value group",
					type: "button",
					onClick: ($event) => unref(copyToClipboard)(unref(formatExample)(unwrapExampleObject(ex)))
				}, [createElementVNode("span", null, toDisplayString(unref(formatExample)(unwrapExampleObject(ex))), 1), createVNode(unref(ScalarIcon), {
					class: "text-c-3 group-hover:text-c-1 ml-auto min-h-3 min-w-3",
					icon: "Clipboard",
					size: "xs"
				})], 8, _hoisted_5);
			}), 128))])])) : createCommentVNode("", true)], 64);
		};
	}
});
//#endregion
export { SchemaPropertyExamples_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaPropertyExamples.vue.script.js.map