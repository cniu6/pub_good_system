import { useLocalization } from "../../features/localization/use-localization.js";
import ScreenReader_default from "../ScreenReader.vue.js";
import OperationsListItem_default from "./OperationsListItem.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
import { ScalarCard, ScalarCardHeader, ScalarCardSection } from "@scalar/components/card";
//#region src/components/OperationsList/OperationsList.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var OperationsList_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationsList",
	props: {
		tag: {},
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const operationsAndWebhooks = computed(() => {
			return __props.tag.children?.filter((child) => child.type === "operation" || child.type === "webhook") ?? [];
		});
		return (_ctx, _cache) => {
			return __props.tag.children && __props.tag.children?.length > 0 ? (openBlock(), createBlock(unref(ScalarCard), {
				key: 0,
				class: "endpoints-card"
			}, {
				default: withCtx(() => [createVNode(unref(ScalarCardHeader), { muted: "" }, {
					default: withCtx(() => [createVNode(ScreenReader_default, null, {
						default: withCtx(() => [createTextVNode(toDisplayString(__props.tag.title), 1)]),
						_: 1
					}), createTextVNode(" " + toDisplayString(__props.tag.isWebhooks ? unref(translate)("navigation.webhooks") : unref(translate)("navigation.operations")), 1)]),
					_: 1
				}), createVNode(unref(ScalarCardSection), { class: "custom-scroll max-h-[60vh]" }, {
					default: withCtx(() => [createElementVNode("ul", {
						"aria-label": unref(translate)("navigation.endpoints", { name: __props.tag.title }),
						class: "endpoints"
					}, [(openBlock(true), createElementBlock(Fragment, null, renderList(operationsAndWebhooks.value, (operationOrWebhook) => {
						return openBlock(), createBlock(OperationsListItem_default, {
							key: operationOrWebhook.id,
							eventBus: __props.eventBus,
							operation: operationOrWebhook
						}, null, 8, ["eventBus", "operation"]);
					}), 128))], 8, _hoisted_1)]),
					_: 1
				})]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { OperationsList_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationsList.vue.script.js.map