import { Fragment, createBlock, createCommentVNode, createElementBlock, createTextVNode, defineComponent, openBlock, renderList, resolveDynamicComponent, toDisplayString, withCtx } from "vue";
//#region src/features/Operation/components/SecurityRequirementBadgeScheme.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { key: 0 };
var _hoisted_2 = { key: 1 };
var SecurityRequirementBadgeScheme_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SecurityRequirementBadgeScheme",
	props: {
		is: { default: "li" },
		scheme: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(resolveDynamicComponent(__props.is), { class: "markdown" }, {
				default: withCtx(() => [
					createTextVNode(toDisplayString(__props.scheme.name) + " ", 1),
					__props.scheme.scheme?.type ? (openBlock(), createElementBlock("code", _hoisted_1, toDisplayString(__props.scheme.scheme.type), 1)) : createCommentVNode("", true),
					__props.scheme.scopes.length ? (openBlock(), createElementBlock("ul", _hoisted_2, [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.scheme.scopes, (scope) => {
						return openBlock(), createElementBlock("li", {
							key: scope,
							class: "font-code text-c-2"
						}, toDisplayString(scope), 1);
					}), 128))])) : createCommentVNode("", true)
				]),
				_: 1
			});
		};
	}
});
//#endregion
export { SecurityRequirementBadgeScheme_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SecurityRequirementBadgeScheme.vue.script.js.map