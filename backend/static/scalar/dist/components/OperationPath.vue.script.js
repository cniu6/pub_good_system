import { Fragment, computed, createElementBlock, createTextVNode, defineComponent, normalizeClass, openBlock, renderList, toDisplayString } from "vue";
//#region src/components/OperationPath.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { key: 0 };
var OperationPath_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationPath",
	props: {
		path: {},
		deprecated: { type: Boolean }
	},
	setup(__props) {
		const props = __props;
		const isVariable = (part) => part.startsWith("{") && part.endsWith("}");
		const pathParts = computed(() => props.path.split(/({[^}]+})/));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("span", { class: normalizeClass(["operation-path", { deprecated: __props.deprecated }]) }, [(openBlock(true), createElementBlock(Fragment, null, renderList(pathParts.value, (part, i) => {
				return openBlock(), createElementBlock(Fragment, { key: i }, [isVariable(part) ? (openBlock(), createElementBlock("em", _hoisted_1, toDisplayString(part), 1)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(part), 1)], 64))], 64);
			}), 128))], 2);
		};
	}
});
//#endregion
export { OperationPath_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationPath.vue.script.js.map