import { useLocalization } from "../../features/localization/use-localization.js";
import { useAgentContext } from "../../hooks/use-agent.js";
import { Fragment, Transition, createElementBlock, createElementVNode, createVNode, defineAsyncComponent, defineComponent, openBlock, unref, vShow, withCtx, withDirectives, withKeys } from "vue";
import { ScalarIconX } from "@scalar/icons";
import { ScalarIconButton } from "@scalar/components/icon-button";
//#region src/components/AgentScalar/AgentScalarDrawer.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "agent-scalar-container custom-scroll custom-scroll-self-contain-overflow overflow-auto px-6" };
var AgentScalarDrawer_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AgentScalarDrawer",
	props: {
		agentScalarConfiguration: {},
		externalUrls: {},
		workspaceStore: {}
	},
	setup(__props) {
		const agentContext = useAgentContext();
		const { translate } = useLocalization();
		const AgentScalarChatInterface = defineAsyncComponent(async () => import("./AgentScalarChatInterface.vue.js"));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [createVNode(Transition, {
				enterActiveClass: "transition-opacity duration-500",
				enterFromClass: "opacity-0",
				enterToClass: "opacity-100",
				leaveActiveClass: "transition-opacity duration-200",
				leaveFromClass: "opacity-100",
				leaveToClass: "opacity-0"
			}, {
				default: withCtx(() => [withDirectives(createElementVNode("div", {
					class: "agent-scalar-overlay bg-backdrop fixed inset-0 z-10 ease-[cubic-bezier(0.77,0,0.175,1)]",
					onClick: _cache[0] || (_cache[0] = ($event) => unref(agentContext)?.closeAgent())
				}, null, 512), [[vShow, unref(agentContext)?.showAgent.value]])]),
				_: 1
			}), createVNode(Transition, {
				enterActiveClass: "transition-transform duration-300",
				enterFromClass: "-translate-x-full",
				enterToClass: "translate-x-0",
				leaveActiveClass: "transition-transform duration-200",
				leaveFromClass: "translate-x-0",
				leaveToClass: "-translate-x-full"
			}, {
				default: withCtx(() => [withDirectives(createElementVNode("div", {
					class: "agent-scalar left-refs-w-sidebar bg-b-1 fixed inset-y-0 right-12 z-10 grid border-r shadow-lg",
					onKeydown: _cache[2] || (_cache[2] = withKeys(($event) => unref(agentContext)?.closeAgent(), ["escape"]))
				}, [createElementVNode("div", _hoisted_1, [createVNode(unref(AgentScalarChatInterface), {
					agentScalarConfiguration: __props.agentScalarConfiguration,
					externalUrls: __props.externalUrls,
					prefilledMessage: unref(agentContext)?.prefilledMessage,
					workspaceStore: __props.workspaceStore
				}, null, 8, [
					"agentScalarConfiguration",
					"externalUrls",
					"prefilledMessage",
					"workspaceStore"
				])]), createVNode(unref(ScalarIconButton), {
					class: "agent-scalar-exit-button absolute top-2 right-2",
					icon: unref(ScalarIconX),
					label: unref(translate)("agent.close"),
					weight: "bold",
					onClick: _cache[1] || (_cache[1] = ($event) => unref(agentContext)?.closeAgent())
				}, null, 8, ["icon", "label"])], 544), [[vShow, unref(agentContext)?.showAgent.value]])]),
				_: 1
			})], 64);
		};
	}
});
//#endregion
export { AgentScalarDrawer_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AgentScalarDrawer.vue.script.js.map