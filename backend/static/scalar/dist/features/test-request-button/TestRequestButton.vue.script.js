import { useLocalization } from "../localization/use-localization.js";
import ScreenReader_default from "../../components/ScreenReader.vue.js";
import { createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref, withCtx, withModifiers } from "vue";
import { ScalarIconPlay } from "@scalar/icons";
//#region src/features/test-request-button/TestRequestButton.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["method"];
var TestRequestButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "TestRequestButton",
	props: {
		id: {},
		method: {},
		path: {},
		eventBus: {},
		exampleName: {},
		requestBodyCompositionSelection: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/** Route via ID and optionally with example name */
		const handleClick = () => {
			const payload = {
				id: __props.id,
				...__props.exampleName && { exampleName: __props.exampleName },
				...__props.requestBodyCompositionSelection && Object.keys(__props.requestBodyCompositionSelection).length > 0 && { requestBodyCompositionSelection: __props.requestBodyCompositionSelection }
			};
			__props.eventBus.emit("ui:open:client-modal", payload);
		};
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("button", {
				class: "show-api-client-button",
				method: __props.method,
				type: "button",
				onClick: withModifiers(handleClick, ["stop"])
			}, [
				createVNode(unref(ScalarIconPlay), {
					class: "size-3",
					weight: "fill"
				}),
				createElementVNode("span", null, toDisplayString(unref(translate)("operation.testRequest")), 1),
				createVNode(ScreenReader_default, null, {
					default: withCtx(() => [createTextVNode("(" + toDisplayString(__props.method) + " " + toDisplayString(__props.path) + ")", 1)]),
					_: 1
				})
			], 8, _hoisted_1);
		};
	}
});
//#endregion
export { TestRequestButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=TestRequestButton.vue.script.js.map