import HttpMethod_default from "../../../../components/HttpMethod/HttpMethod.vue.js";
import OperationParameters_default from "../OperationParameters.vue.js";
import OperationResponses_default from "../OperationResponses.vue.js";
import { createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconCaretRight } from "@scalar/icons";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/features/Operation/components/callbacks/Callback.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "group callback-list-item" };
var _hoisted_2 = { class: "font-code bg-b-1 callback-sticky-offset callback-list-item-title sticky flex cursor-pointer flex-row items-start gap-2 border-t py-2.5 text-sm group-open:flex-wrap" };
var _hoisted_3 = { class: "text-c-1 min-w-0 flex-1 truncate text-sm leading-5 font-bold group-open:whitespace-normal" };
var _hoisted_4 = { class: "text-c-2 font-normal" };
var _hoisted_5 = { class: "callback-operation-container flex flex-col gap-2" };
var Callback_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Callback",
	props: {
		callback: {},
		method: {},
		name: {},
		url: {},
		eventBus: {},
		document: {},
		options: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("details", _hoisted_1, [createElementVNode("summary", _hoisted_2, [
				createVNode(unref(ScalarIconCaretRight), {
					class: "callback-list-item-icon text-c-3 group-hover:text-c-1 absolute top-3.5 -left-5 size-3 transition-transform duration-100 group-open:rotate-90",
					weight: "bold"
				}),
				createVNode(unref(HttpMethod_default), {
					as: "span",
					class: "request-method py-0.75 font-bold",
					method: __props.method
				}, null, 8, ["method"]),
				createElementVNode("div", _hoisted_3, [createTextVNode(toDisplayString(__props.name) + " ", 1), createElementVNode("span", _hoisted_4, toDisplayString(__props.url), 1)])
			]), createElementVNode("div", _hoisted_5, [createVNode(OperationParameters_default, {
				document: __props.document,
				eventBus: __props.eventBus,
				options: __props.options,
				parameters: __props.callback.parameters?.map((param) => unref(getResolvedRef)(param)) ?? [],
				requestBody: unref(getResolvedRef)(__props.callback.requestBody)
			}, null, 8, [
				"document",
				"eventBus",
				"options",
				"parameters",
				"requestBody"
			]), createVNode(OperationResponses_default, {
				collapsableItems: false,
				document: __props.document,
				eventBus: __props.eventBus,
				options: __props.options,
				responses: __props.callback.responses
			}, null, 8, [
				"document",
				"eventBus",
				"options",
				"responses"
			])])]);
		};
	}
});
//#endregion
export { Callback_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Callback.vue.script.js.map