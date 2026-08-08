import { useLocalization } from "../../localization/use-localization.js";
import { getRequiredScopeGroups } from "../helpers/get-required-security.js";
import { Fragment, computed, createCommentVNode, createElementBlock, createElementVNode, defineComponent, normalizeClass, openBlock, renderList, toDisplayString, unref } from "vue";
//#region src/features/Operation/components/OperationScopes.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "mt-6"
};
var _hoisted_2 = { class: "text-c-1 mt-3 mb-3 text-lg leading-[1.45] font-medium" };
var _hoisted_3 = {
	key: 0,
	class: "text-c-2 mb-2 text-sm"
};
var OperationScopes_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationScopes",
	props: { requiredSecurity: {} },
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* Required OAuth scopes grouped by security alternative (OR). A single group is the
		* common case and renders as a plain list. Multiple groups are kept separate so that
		* mutually exclusive scope sets are not implied to be required all at once.
		*/
		const scopeGroups = computed(() => getRequiredScopeGroups(__props.requiredSecurity));
		/**
		* Whether at least one security alternative is satisfied without any OAuth scopes
		* (for example an API key or HTTP bearer). Those alternatives carry no scopes, so they
		* are absent from `scopeGroups`, but their existence means the listed scopes are only
		* required for some auth paths — not mandatory for the operation.
		*/
		const hasScopeFreeAlternative = computed(() => __props.requiredSecurity.requirements.some((group) => group.schemes.every((scheme) => scheme.scopes.length === 0)));
		/**
		* Show the "one of" hint whenever the listed scopes are just one of several auth
		* alternatives — either multiple scoped groups, or a single scoped group alongside a
		* scope-free alternative. Without it, a lone scope list would read as mandatory.
		*/
		const showAlternativesHint = computed(() => scopeGroups.value.length > 1 || scopeGroups.value.length > 0 && hasScopeFreeAlternative.value);
		return (_ctx, _cache) => {
			return scopeGroups.value.length ? (openBlock(), createElementBlock("div", _hoisted_1, [
				createElementVNode("div", _hoisted_2, toDisplayString(unref(translate)("authentication.scopes")), 1),
				showAlternativesHint.value ? (openBlock(), createElementBlock("div", _hoisted_3, toDisplayString(unref(translate)("authentication.oneOf")), 1)) : createCommentVNode("", true),
				(openBlock(true), createElementBlock(Fragment, null, renderList(scopeGroups.value, (group, index) => {
					return openBlock(), createElementBlock("ul", {
						key: index,
						class: normalizeClass(["mb-3 list-none p-0 text-sm", { "mt-3 border-t pt-3": scopeGroups.value.length > 1 && index > 0 }])
					}, [(openBlock(true), createElementBlock(Fragment, null, renderList(group, (scope) => {
						return openBlock(), createElementBlock("li", {
							key: scope,
							class: "font-code text-c-2"
						}, toDisplayString(scope), 1);
					}), 128))], 2);
				}), 128))
			])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { OperationScopes_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationScopes.vue.script.js.map