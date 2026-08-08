import { useLocalization } from "../../../features/localization/use-localization.js";
import Selector_default from "./Selector.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref, useId } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ServerVariablesForm } from "@scalar/api-client/components/Server";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/blocks/scalar-asyncapi-server-selector-block/components/AsyncApiServerSelector.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "bg-b-2 flex h-8 items-center rounded-t-xl border-x border-t px-3 py-2.5 font-medium" };
var _hoisted_2 = ["id"];
var AsyncApiServerSelector_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AsyncApiServerSelector",
	props: {
		eventBus: {},
		selectedServer: {},
		servers: {}
	},
	setup(__props) {
		const id = useId();
		const { translate } = useLocalization();
		/**
		* Normalize AsyncAPI server variables into the shape the shared
		* ServerVariablesForm expects (resolving references and defaulting `default`).
		*/
		const serverVariables = computed(() => {
			const variables = __props.selectedServer?.server.variables;
			if (!variables) return;
			return Object.fromEntries(Object.entries(variables).map(([name, variable]) => {
				const resolved = getResolvedRef(variable);
				return [name, {
					default: resolved.default ?? "",
					enum: resolved.enum,
					description: resolved.description
				}];
			}));
		});
		/** Update the selected server */
		const updateServer = (name) => {
			__props.eventBus.emit("asyncapi-server:update:selected", { name });
		};
		/** Update a server variable on the selected server */
		const updateServerVariable = (key, value) => {
			if (!__props.selectedServer) return;
			__props.eventBus.emit("asyncapi-server:update:variables", {
				name: __props.selectedServer.name,
				key,
				value
			});
		};
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [
				createElementVNode("label", _hoisted_1, toDisplayString(unref(translate)("server.label")), 1),
				createElementVNode("div", {
					id: unref(id),
					class: normalizeClass(["border", { "rounded-b-xl": !__props.selectedServer?.description && !serverVariables.value }])
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
					variables: serverVariables.value,
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
export { AsyncApiServerSelector_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AsyncApiServerSelector.vue.script.js.map