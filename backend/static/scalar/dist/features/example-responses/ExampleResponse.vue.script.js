import { useLocalization } from "../localization/use-localization.js";
import { computed, createBlock, createElementBlock, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { getExampleFromSchema } from "@scalar/workspace-store/request-example";
import { getResolvedRefDeep } from "@scalar/blocks/code-example";
import { ScalarCodeBlock } from "@scalar/components/code-block";
import { ScalarVirtualCodeBlock } from "@scalar/components/virtual-code-block";
import { prettyPrintJson } from "@scalar/helpers/json/pretty-print-json";
//#region src/features/example-responses/ExampleResponse.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 2,
	class: "empty-state"
};
var VIRTUALIZATION_THRESHOLD = 2e4;
var ExampleResponse_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExampleResponse",
	props: {
		response: {},
		example: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/** Get content from the appropriate source */
		const getContent = () => {
			if (__props.example !== void 0) return getResolvedRefDeep(__props.example)?.value ?? "";
			if (__props.response?.schema) return getExampleFromSchema(getResolvedRefDeep(__props.response.schema), {
				emptyString: "string",
				mode: "read"
			});
		};
		/** Pre-pretty printed content string, avoids multiple pretty prints*/
		const prettyPrintedContent = computed(() => {
			const content = getContent();
			if (content === void 0) return;
			return prettyPrintJson(content);
		});
		const shouldVirtualize = computed(() => {
			if (prettyPrintedContent.value === void 0) return false;
			return prettyPrintedContent.value.length > VIRTUALIZATION_THRESHOLD;
		});
		return (_ctx, _cache) => {
			return prettyPrintedContent.value !== void 0 && !shouldVirtualize.value ? (openBlock(), createBlock(unref(ScalarCodeBlock), {
				key: 0,
				class: "bg-b-2",
				lang: "json",
				prettyPrintedContent: prettyPrintedContent.value
			}, null, 8, ["prettyPrintedContent"])) : prettyPrintedContent.value !== void 0 && shouldVirtualize.value ? (openBlock(), createBlock(unref(ScalarVirtualCodeBlock), {
				key: 1,
				class: "bg-b-2",
				content: prettyPrintedContent.value,
				lang: "json"
			}, null, 8, ["content"])) : (openBlock(), createElementBlock("div", _hoisted_1, toDisplayString(unref(translate)("response.noBody")), 1));
		};
	}
});
//#endregion
export { ExampleResponse_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExampleResponse.vue.script.js.map