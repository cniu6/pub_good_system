import { useLocalization } from "../../features/localization/use-localization.js";
import { useAgentContext } from "../../hooks/use-agent.js";
import { createElementBlock, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconSparkle } from "@scalar/icons";
//#region src/components/AgentScalar/AgentScalarButton.vue?vue&type=script&setup=true&lang.ts
var AgentScalarButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AgentScalarButton",
	setup(__props) {
		const agentContext = useAgentContext();
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("button", {
				class: "bg-sidebar-b-search text-sidebar-c-2 hover:text-sidebar-c-1 flex items-center gap-1.5 rounded border px-2 text-base whitespace-nowrap",
				type: "button",
				onClick: _cache[0] || (_cache[0] = ($event) => unref(agentContext)?.toggleAgent())
			}, [createVNode(unref(ScalarIconSparkle)), createTextVNode(" " + toDisplayString(unref(translate)("agent.askAi")), 1)]);
		};
	}
});
//#endregion
export { AgentScalarButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AgentScalarButton.vue.script.js.map