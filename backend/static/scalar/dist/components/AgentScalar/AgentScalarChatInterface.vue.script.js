import { createBlock, defineComponent, openBlock, unref } from "vue";
import { Chat } from "@scalar/agent-chat";
//#region src/components/AgentScalar/AgentScalarChatInterface.vue?vue&type=script&setup=true&lang.ts
var AgentScalarChatInterface_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AgentScalarChatInterface",
	props: {
		agentScalarConfiguration: {},
		externalUrls: {},
		workspaceStore: {},
		prefilledMessage: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(Chat), {
				baseUrl: __props.externalUrls.apiBaseUrl,
				dashboardUrl: __props.externalUrls.dashboardUrl,
				getActiveDocumentJson: () => __props.workspaceStore.exportActiveDocument("json"),
				getAgentKey: __props.agentScalarConfiguration?.key ? () => __props.agentScalarConfiguration?.key ?? "" : void 0,
				hideAddApi: __props.agentScalarConfiguration?.hideAddApi,
				mode: __props.agentScalarConfiguration?.key ? "full" : "preview",
				platformProxyUrl: __props.externalUrls.proxyUrl,
				prefilledMessage: __props.prefilledMessage,
				registryDocuments: [],
				registryUrl: __props.externalUrls.registryUrl
			}, null, 8, [
				"baseUrl",
				"dashboardUrl",
				"getActiveDocumentJson",
				"getAgentKey",
				"hideAddApi",
				"mode",
				"platformProxyUrl",
				"prefilledMessage",
				"registryUrl"
			]);
		};
	}
});
//#endregion
export { AgentScalarChatInterface_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AgentScalarChatInterface.vue.script.js.map