import { getRequiredSecurity } from "./helpers/get-required-security.js";
import { filterSelectedSecurity } from "./helpers/filter-selected-security.js";
import ClassicLayout_default from "./layouts/ClassicLayout.vue.js";
import ModernLayout_default from "./layouts/ModernLayout.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock } from "vue";
import { combineParams } from "@scalar/workspace-store/request-example";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { getFirstServer } from "@scalar/workspace-store/helpers/get-first-server";
var Operation_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Operation",
	props: {
		id: {},
		method: {},
		options: {},
		document: {},
		path: {},
		pathValue: {},
		server: {},
		securitySchemes: {},
		clientOptions: {},
		isCollapsed: { type: Boolean },
		isWebhook: { type: Boolean },
		selectedClient: {},
		selectedExample: {},
		eventBus: {},
		authStore: {}
	},
	setup(__props) {
		/**
		* Operation from the new workspace store, ensure we are de-reference
		*
		* Also adds in params from the pathItemObject
		*/
		const operation = computed(() => {
			const entity = getResolvedRef(__props.pathValue?.[__props.method]);
			if (!entity) return null;
			const parameters = combineParams(__props.pathValue?.parameters, entity.parameters);
			return {
				...entity,
				parameters
			};
		});
		/**
		* Determine the effective server for the code examples.
		*/
		const selectedServer = computed(() => getFirstServer(operation.value?.servers ?? null, __props.pathValue?.servers ?? null, __props.server));
		const requiredSecurity = computed(() => getRequiredSecurity(operation.value, __props.document));
		/** We must ensure the selected security schemes are required on this operation */
		const selectedSecuritySchemes = computed(() => filterSelectedSecurity(__props.document, operation.value, __props.authStore.getAuthSelectedSchemas({
			type: "document",
			documentName: __props.document?.["x-scalar-navigation"]?.name ?? ""
		}), __props.authStore.getAuthSelectedSchemas({
			type: "operation",
			documentName: __props.document?.["x-scalar-navigation"]?.name ?? "",
			path: __props.path,
			method: __props.method
		}), __props.securitySchemes));
		return (_ctx, _cache) => {
			return operation.value ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [__props.options.layout === "classic" ? (openBlock(), createBlock(ClassicLayout_default, {
				key: 0,
				id: __props.id,
				clientOptions: __props.clientOptions,
				document: __props.document,
				eventBus: __props.eventBus,
				isCollapsed: __props.isCollapsed,
				isWebhook: __props.isWebhook,
				method: __props.method,
				operation: operation.value,
				options: __props.options,
				path: __props.path,
				requiredSecurity: requiredSecurity.value,
				selectedClient: __props.selectedClient,
				selectedExample: __props.selectedExample,
				selectedSecuritySchemes: selectedSecuritySchemes.value,
				selectedServer: selectedServer.value
			}, null, 8, [
				"id",
				"clientOptions",
				"document",
				"eventBus",
				"isCollapsed",
				"isWebhook",
				"method",
				"operation",
				"options",
				"path",
				"requiredSecurity",
				"selectedClient",
				"selectedExample",
				"selectedSecuritySchemes",
				"selectedServer"
			])) : (openBlock(), createBlock(ModernLayout_default, {
				key: 1,
				id: __props.id,
				clientOptions: __props.clientOptions,
				document: __props.document,
				eventBus: __props.eventBus,
				isWebhook: __props.isWebhook,
				method: __props.method,
				operation: operation.value,
				options: __props.options,
				path: __props.path,
				requiredSecurity: requiredSecurity.value,
				selectedClient: __props.selectedClient,
				selectedExample: __props.selectedExample,
				selectedSecuritySchemes: selectedSecuritySchemes.value,
				selectedServer: selectedServer.value
			}, null, 8, [
				"id",
				"clientOptions",
				"document",
				"eventBus",
				"isWebhook",
				"method",
				"operation",
				"options",
				"path",
				"requiredSecurity",
				"selectedClient",
				"selectedExample",
				"selectedSecuritySchemes",
				"selectedServer"
			]))], 64)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Operation_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Operation.vue.script.js.map