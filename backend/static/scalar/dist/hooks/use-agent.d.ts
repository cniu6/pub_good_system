import type { ComputedRef, Ref } from 'vue';
import { type InjectionKey } from 'vue';
type UseAgentOptions = {
    /** Optional. When provided, controls whether the agent UI is enabled (e.g. from doc config). Defaults to isLocalUrl. */
    agentEnabled?: ComputedRef<boolean>;
};
type UseAgentReturn = {
    showAgent: Ref<boolean>;
    agentEnabled: ComputedRef<boolean>;
    /** Ref used to pass a prefill message into the agent when opening. Cleared on close. */
    prefilledMessage: Ref<string>;
    openAgent: (message?: string) => void;
    closeAgent: () => void;
    toggleAgent: () => void;
};
export declare const AGENT_CONTEXT_SYMBOL: InjectionKey<UseAgentReturn>;
/**
 * Hook for agent visibility and enabled state.
 * Call from the API Reference root (e.g. ApiReference.vue) with options to create the state, then provide it so descendants can inject it.
 *
 * Returns:
 * - showAgent: whether the agent panel is visible
 * - agentEnabled: whether the agent is enabled
 * - openAgent, closeAgent, toggleAgent: imperative controls
 */
export declare function useAgent(options: UseAgentOptions): UseAgentReturn;
/**
 * Inject the agent context provided by ApiReference.
 * Falls back to module-level state when inject is undefined (e.g. async boundary or mount order).
 * Use in descendant components (e.g. AskAgentButton) to open the agent or check if it is enabled.
 *
 * Returns a computed ref; use v-if="agentContext?.agentEnabled" so the button only renders when context exists and agent is enabled.
 */
export declare function useAgentContext(): ComputedRef<UseAgentReturn | undefined>;
export {};
//# sourceMappingURL=use-agent.d.ts.map