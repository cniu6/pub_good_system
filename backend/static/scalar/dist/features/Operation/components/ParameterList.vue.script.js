import ParameterListItem_default from "./ParameterListItem.vue.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createElementVNode, defineComponent, openBlock, renderList, renderSlot, unref, useId } from "vue";
//#region src/features/Operation/components/ParameterList.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "mt-6"
};
var _hoisted_2 = ["id"];
var _hoisted_3 = ["aria-labelledby"];
var ParameterList_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ParameterList",
	props: {
		parameters: {},
		breadcrumb: {},
		eventBus: {},
		collapsableItems: { type: Boolean },
		document: {},
		options: {}
	},
	setup(__props) {
		/** Accessible id for the heading */
		const id = useId();
		return (_ctx, _cache) => {
			return __props.parameters?.length ? (openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("div", {
				id: unref(id),
				class: "text-c-1 mt-3 mb-3 text-lg leading-[1.45] font-medium"
			}, [renderSlot(_ctx.$slots, "title")], 8, _hoisted_2), createElementVNode("ul", {
				"aria-labelledby": unref(id),
				class: "mb-3 list-none p-0 text-sm"
			}, [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.parameters, (item) => {
				return openBlock(), createBlock(ParameterListItem_default, {
					key: item.name,
					breadcrumb: __props.breadcrumb,
					collapsableItems: __props.collapsableItems,
					document: __props.document,
					eventBus: __props.eventBus,
					name: item.name,
					options: __props.options,
					parameter: item
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
export { ParameterList_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ParameterList.vue.script.js.map