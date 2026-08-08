import { useLocalization } from "../localization/use-localization.js";
import { useAgentContext } from "../../hooks/use-agent.js";
import { createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, ref, toDisplayString, unref, vModelText, withDirectives, withModifiers } from "vue";
import { ScalarIconArrowUp, ScalarIconSparkle } from "@scalar/icons";
//#region src/features/ask-agent-button/AskAgentButton.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "ask-agent-scalar-input-label" };
var _hoisted_2 = ["placeholder"];
var _hoisted_3 = {
	class: "ask-agent-scalar-send",
	type: "submit"
};
var AskAgentButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AskAgentButton",
	setup(__props) {
		const agentContext = useAgentContext();
		const { translate } = useLocalization();
		const message = ref("");
		const inputRef = ref();
		function handleSubmit() {
			agentContext.value?.openAgent(message.value);
			message.value = "";
		}
		return (_ctx, _cache) => {
			return unref(agentContext)?.agentEnabled.value ? (openBlock(), createElementBlock("form", {
				key: 0,
				class: "agent-button-container",
				onClick: _cache[1] || (_cache[1] = ($event) => inputRef.value?.focus()),
				onSubmit: _cache[2] || (_cache[2] = withModifiers(($event) => handleSubmit(), ["prevent"]))
			}, [
				createVNode(unref(ScalarIconSparkle), {
					class: "size-3 shrink-0",
					weight: "fill"
				}),
				createElementVNode("div", _hoisted_1, toDisplayString(unref(translate)("agent.askAiAgent")), 1),
				withDirectives(createElementVNode("input", {
					ref_key: "inputRef",
					ref: inputRef,
					"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => message.value = $event),
					class: normalizeClass(["ask-agent-scalar-input", { "ask-agent-scalar-input-not-empty": message.value.length > 0 }]),
					placeholder: unref(translate)("agent.askAiAgent")
				}, null, 10, _hoisted_2), [[vModelText, message.value]]),
				createElementVNode("button", _hoisted_3, [createVNode(unref(ScalarIconArrowUp), {
					class: "size-3",
					weight: "bold"
				})])
			], 32)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { AskAgentButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AskAgentButton.vue.script.js.map