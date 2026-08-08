import { computed, createBlock, defineComponent, openBlock, unref } from "vue";
import { getResolvedRefDeep } from "@scalar/blocks/code-example";
import { ScalarCodeBlock } from "@scalar/components/code-block";
import { ScalarVirtualCodeBlock } from "@scalar/components/virtual-code-block";
import { prettyPrintJson } from "@scalar/helpers/json/pretty-print-json";
//#region src/features/example-responses/ExampleSchema.vue?vue&type=script&setup=true&lang.ts
var VIRTUALIZATION_THRESHOLD = 2e4;
var ExampleSchema_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExampleSchema",
	props: {
		id: {},
		schema: {}
	},
	setup(__props) {
		const schemaContent = computed(() => {
			if (!__props.schema) return;
			return prettyPrintJson(getResolvedRefDeep(__props.schema));
		});
		const shouldVirtualizeSchema = computed(() => {
			return (schemaContent.value?.length ?? 0) > VIRTUALIZATION_THRESHOLD;
		});
		return (_ctx, _cache) => {
			return !shouldVirtualizeSchema.value ? (openBlock(), createBlock(unref(ScalarCodeBlock), {
				key: 0,
				id: __props.id,
				class: "bg-b-2",
				lang: "json",
				prettyPrintedContent: schemaContent.value ?? ""
			}, null, 8, ["id", "prettyPrintedContent"])) : (openBlock(), createBlock(unref(ScalarVirtualCodeBlock), {
				key: 1,
				id: __props.id,
				class: "bg-b-2",
				content: schemaContent.value ?? "",
				lang: "json"
			}, null, 8, ["id", "content"]));
		};
	}
});
//#endregion
export { ExampleSchema_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExampleSchema.vue.script.js.map