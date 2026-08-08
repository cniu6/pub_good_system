import { useLocalization } from "../../localization/use-localization.js";
import ParameterListItem_default from "./ParameterListItem.vue.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createElementVNode, defineComponent, openBlock, renderList, toDisplayString, unref } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/features/Operation/components/OperationResponses.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "mt-6"
};
var _hoisted_2 = { class: "text-c-1 mt-3 mb-3 leading-[1.45] font-medium" };
var _hoisted_3 = ["aria-label"];
var OperationResponses_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationResponses",
	props: {
		responses: {},
		breadcrumb: {},
		collapsableItems: { type: Boolean },
		eventBus: {},
		document: {},
		options: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return Object.keys(__props.responses ?? {}).length ? (openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("div", _hoisted_2, toDisplayString(unref(translate)("operation.responses")), 1), createElementVNode("ul", {
				"aria-label": unref(translate)("operation.responses"),
				class: "mb-3 list-none p-0 text-sm"
			}, [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.responses, (response, status) => {
				return openBlock(), createBlock(ParameterListItem_default, {
					key: status,
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "responses"] : void 0,
					collapsableItems: __props.collapsableItems,
					document: __props.document,
					eventBus: __props.eventBus,
					name: status,
					options: __props.options,
					parameter: unref(getResolvedRef)(response)
				}, null, 8, [
					"breadcrumb",
					"collapsableItems",
					"document",
					"eventBus",
					"name",
					"options",
					"parameter"
				]);
			}), 128))], 8, _hoisted_3)])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { OperationResponses_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationResponses.vue.script.js.map