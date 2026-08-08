import { computed, defineComponent, toDisplayString } from "vue";
//#region src/components/Content/Schema/RenderString.vue?vue&type=script&setup=true&lang.ts
var RenderString_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "RenderString",
	props: { value: {} },
	setup(__props) {
		/**
		* Give this component any data, and it tries to render a meaningful string.
		*
		* @example
		* <RenderString :value="1" /> => "1"
		* <RenderString :value="true" /> => "true"
		* <RenderString :value="false" /> => "false"
		* <RenderString :value="null" /> => "null"
		* <RenderString :value="undefined" /> => "undefined"
		* <RenderString :value="{}" /> => "{}"
		* <RenderString :value="[]" /> => "[]"
		* <RenderString :value="['a', 'b', 'c']" /> => "['a', 'b', 'c']"
		* <RenderString :value="() => 'hello'" /> => "() => 'hello'"
		* <RenderString :value="Symbol('test')" /> => "Symbol(test)"
		**/
		const valueAsString = computed(() => {
			if (__props.value === "") return `''`;
			if (__props.value === null) return "null";
			if (__props.value === void 0) return "undefined";
			return __props.value;
		});
		return (_ctx, _cache) => {
			return toDisplayString(valueAsString.value);
		};
	}
});
//#endregion
export { RenderString_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=RenderString.vue.script.js.map