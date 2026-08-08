import { computed, createBlock, createCommentVNode, createElementVNode, createVNode, defineComponent, openBlock, ref, unref, withCtx } from "vue";
import { ExamplePicker } from "@scalar/blocks/code-example";
import { ScalarCard, ScalarCardFooter, ScalarCardSection } from "@scalar/components/card";
import { ScalarCodeBlock } from "@scalar/components/code-block";
//#region src/components/Content/Schema/SchemaObjectExampleCodeBlock.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "code-snippet" };
var SchemaObjectExampleCodeBlock_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaObjectExampleCodeBlock",
	props: { schema: {} },
	setup(__props) {
		/** Grab the examples for the schema object */
		const examples = computed(() => {
			const base = {};
			if (__props.schema["x-examples"]) Object.entries(__props.schema["x-examples"]).forEach(([key, value]) => {
				base[key] = typeof value === "object" ? JSON.stringify(value, null, 2) : String(value);
			});
			if (__props.schema.examples) __props.schema.examples.forEach((value, index) => {
				base[`Example ${index + 1}`] = typeof value === "object" ? JSON.stringify(value, null, 2) : String(value);
			});
			return base;
		});
		/** The currently selected example key */
		const selectedExampleKey = ref(Object.keys(examples.value)[0] ?? "");
		return (_ctx, _cache) => {
			return Object.keys(examples.value).length > 0 ? (openBlock(), createBlock(unref(ScalarCard), {
				key: 0,
				class: "dark-mode"
			}, {
				default: withCtx(() => [createVNode(unref(ScalarCardSection), null, {
					default: withCtx(() => [createElementVNode("div", _hoisted_1, [createVNode(unref(ScalarCodeBlock), {
						class: "bg-b-2",
						lang: "json",
						lineNumbers: "",
						prettyPrintedContent: examples.value[selectedExampleKey.value] ?? "There was an error loading the example"
					}, null, 8, ["prettyPrintedContent"])])]),
					_: 1
				}), Object.keys(examples.value).length > 1 ? (openBlock(), createBlock(unref(ScalarCardFooter), {
					key: 0,
					class: "bg-b-3"
				}, {
					default: withCtx(() => [createVNode(unref(ExamplePicker), {
						modelValue: selectedExampleKey.value,
						"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => selectedExampleKey.value = $event),
						examples: examples.value
					}, null, 8, ["modelValue", "examples"])]),
					_: 1
				})) : createCommentVNode("", true)]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SchemaObjectExampleCodeBlock_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaObjectExampleCodeBlock.vue.script.js.map