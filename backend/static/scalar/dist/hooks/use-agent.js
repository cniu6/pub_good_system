import { computed, inject, ref } from "vue";
import { isLocalUrl } from "@scalar/helpers/url/is-local-url";
//#region src/hooks/use-agent.ts
var AGENT_CONTEXT_SYMBOL = Symbol();
/**
* Module-level ref so useAgentContext() can resolve context even when inject fails
* (e.g. async component boundary or mount order). Set when ApiReference calls useAgent().
*/
var agentStateRef = ref(null);
/**
* Hook for agent visibility and enabled state.
* Call from the API Reference root (e.g. ApiReference.vue) with options to create the state, then provide it so descendants can inject it.
*
* Returns:
* - showAgent: whether the agent panel is visible
* - agentEnabled: whether the agent is enabled
* - openAgent, closeAgent, toggleAgent: imperative controls
*/
function useAgent(options) {
	const showAgent = ref(false);
	const prefilledMessage = ref("");
	const agentEnabled = options.agentEnabled ?? computed(() => isLocalUrl(window.location.href));
	const openAgent = (message) => {
		prefilledMessage.value = message ?? "";
		showAgent.value = true;
	};
	const closeAgent = () => {
		showAgent.value = false;
		prefilledMessage.value = "";
	};
	const toggleAgent = () => {
		showAgent.value = !showAgent.value;
		if (!showAgent.value) prefilledMessage.value = "";
	};
	const state = {
		showAgent,
		agentEnabled,
		prefilledMessage,
		openAgent,
		closeAgent,
		toggleAgent
	};
	agentStateRef.value = state;
	return state;
}
/**
* Inject the agent context provided by ApiReference.
* Falls back to module-level state when inject is undefined (e.g. async boundary or mount order).
* Use in descendant components (e.g. AskAgentButton) to open the agent or check if it is enabled.
*
* Returns a computed ref; use v-if="agentContext?.agentEnabled" so the button only renders when context exists and agent is enabled.
*/
function useAgentContext() {
	const injected = inject(AGENT_CONTEXT_SYMBOL, void 0);
	return computed(() => injected ?? agentStateRef.value ?? void 0);
}
//#endregion
export { AGENT_CONTEXT_SYMBOL, useAgent, useAgentContext };

//# sourceMappingURL=use-agent.js.map