import { useLocalization } from "../../../features/localization/use-localization.js";
import Selector_default from "./Selector.vue.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref, useId } from "vue";
import { ServerVariablesForm } from "@scalar/api-client/components/Server";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/blocks/scalar-server-selector-block/components/ServerSelector.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "bg-b-2 flex h-8 items-center rounded-t-xl border-x border-t px-3 py-2.5 font-medium" };
var _hoisted_2 = ["id"];
var ServerSelector_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ServerSelector",
	props: {
		eventBus: {},
		selectedServer: {},
		servers: {}
	},
	setup(__props) {
		const id = useId();
		const { translate } = useLocalization();
		/** Update the selected server */
		const updateServer = (newServer) => {
			__props.eventBus.emit("server:update:selected", {
				url: __props.selectedServer?.url === newServer ? "" : newServer,
				meta: { type: "document" }
			});
		};
		/** Update the server variable */
		const updateServerVariable = (key, value) => {
			/** Find the index of the selected server */
			const index = __props.servers.findIndex((s) => s.url === __props.selectedServer?.url);
			if (index === -1) return;
			__props.eventBus.emit("server:update:variables", {
				index,
				key,
				value,
				meta: { type: "document" }
			});
		};
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [
				createElementVNode("label", _hoisted_1, toDisplayString(unref(translate)("server.label")), 1),
				createElementVNode("div", {
					id: unref(id),
					class: normalizeClass(["border", { "rounded-b-xl": !__props.selectedServer?.description && !__props.selectedServer?.variables }])
				}, [__props.servers.length ? (openBlock(), createBlock(Selector_default, {
					key: 0,
					selectedServer: __props.selectedServer,
					servers: __props.servers,
					target: unref(id),
					"onUpdate:modelValue": updateServer
				}, null, 8, [
					"selectedServer",
					"servers",
					"target"
				])) : createCommentVNode("", true)], 10, _hoisted_2),
				createVNode(unref(ServerVariablesForm), {
					layout: "reference",
					variables: __props.selectedServer?.variables,
					"onUpdate:variable": updateServerVariable
				}, null, 8, ["variables"]),
				__props.selectedServer?.description ? (openBlock(), createBlock(unref(ScalarMarkdown), {
					key: 0,
					class: "text-c-3 rounded-b-xl border-x border-b px-3 py-1.5",
					value: __props.selectedServer.description
				}, null, 8, ["value"])) : createCommentVNode("", true)
			], 64);
		};
	}
});
//#endregion
export { ServerSelector_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ServerSelector.vue.script.js.map