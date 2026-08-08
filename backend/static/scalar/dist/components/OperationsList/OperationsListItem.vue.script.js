import SectionHeaderTag_default from "../Section/SectionHeaderTag.vue.js";
import HttpMethod_default from "../HttpMethod/HttpMethod.vue.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, normalizeStyle, openBlock, toDisplayString, unref, withCtx, withModifiers } from "vue";
import { ScalarIconWebhooksLogo } from "@scalar/icons";
import { getHttpMethodInfo } from "@scalar/helpers/http/http-info";
//#region src/components/OperationsList/OperationsListItem.vue?vue&type=script&setup=true&lang.ts
var OperationsListItem_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationsListItem",
	props: {
		operation: {},
		isCollapsed: { type: Boolean },
		eventBus: {}
	},
	setup(__props) {
		const pathOrTitle = computed(() => {
			if ("path" in __props.operation) return __props.operation.path;
			return __props.operation.title;
		});
		const isWebhook = (_operation) => _operation.type === "webhook";
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("li", {
				key: __props.operation.id,
				class: "contents"
			}, [__props.isCollapsed ? (openBlock(), createBlock(unref(SectionHeaderTag_default), {
				key: 0,
				class: "sr-only",
				level: 3
			}, {
				default: withCtx(() => [createTextVNode(toDisplayString(__props.operation.title) + " (Hidden) ", 1)]),
				_: 1
			})) : createCommentVNode("", true), createElementVNode("a", {
				class: "endpoint",
				onClick: _cache[0] || (_cache[0] = withModifiers(() => __props.eventBus?.emit("scroll-to:nav-item", { id: __props.operation.id }), ["prevent"]))
			}, [createVNode(unref(HttpMethod_default), {
				class: "endpoint-method items-center justify-end gap-2",
				method: __props.operation.method
			}, {
				default: withCtx(() => [isWebhook(__props.operation) ? (openBlock(), createBlock(unref(ScalarIconWebhooksLogo), {
					key: 0,
					class: "size-3.5",
					style: normalizeStyle({ color: unref(getHttpMethodInfo)(__props.operation.method).colorVar })
				}, null, 8, ["style"])) : createCommentVNode("", true)]),
				_: 1
			}, 8, ["method"]), createElementVNode("span", { class: normalizeClass(["endpoint-path", { deprecated: __props.operation.isDeprecated }]) }, toDisplayString(pathOrTitle.value), 3)])]);
		};
	}
});
//#endregion
export { OperationsListItem_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationsListItem.vue.script.js.map