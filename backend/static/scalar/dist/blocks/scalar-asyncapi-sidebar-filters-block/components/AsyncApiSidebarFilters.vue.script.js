import SidebarFilter_default from "./SidebarFilter.vue.js";
import { computed, createBlock, createCommentVNode, createTextVNode, defineComponent, mergeModels, openBlock, unref, useModel, withCtx } from "vue";
import { ScalarSidebarSection } from "@scalar/components/sidebar";
import { getAsyncApiProtocols, getAsyncApiServerOptions } from "@scalar/workspace-store/channel-example";
//#region src/blocks/scalar-asyncapi-sidebar-filters-block/components/AsyncApiSidebarFilters.vue?vue&type=script&setup=true&lang.ts
/**
* AsyncApiSidebarFilters
*
* The "Filters" sidebar section shown for AsyncAPI documents. Bundling the
* native sidebar title and picker pair here keeps the two sidebar layouts
* (modern and classic) from each repeating the option-building logic.
*
* Each picker hides itself when there is nothing to choose from, and the whole
* section hides when neither picker has a real choice — so passing an OpenAPI
* document (or `null`) renders nothing.
*/
var AsyncApiSidebarFilters_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AsyncApiSidebarFilters",
	props: /*@__PURE__*/ mergeModels({
		document: {},
		is: { default: "li" }
	}, {
		"protocol": { default: "" },
		"protocolModifiers": {},
		"server": { default: "" },
		"serverModifiers": {}
	}),
	emits: ["update:protocol", "update:server"],
	setup(__props) {
		/** Selected protocol id; empty string clears the filter. */
		const protocol = useModel(__props, "protocol");
		/** Selected server name; empty string clears the filter. */
		const server = useModel(__props, "server");
		/** Protocol picker options, including the leading "All protocols" entry. */
		const protocolOptions = computed(() => __props.document ? getAsyncApiProtocols(__props.document) : []);
		/** Server picker options, including the leading "All servers" entry. */
		const serverOptions = computed(() => __props.document ? getAsyncApiServerOptions(__props.document) : []);
		/** Each picker is only worth showing when there is a choice beyond "All …". */
		const showProtocol = computed(() => protocolOptions.value.length > 2);
		const showServer = computed(() => serverOptions.value.length > 2);
		return (_ctx, _cache) => {
			return showProtocol.value || showServer.value ? (openBlock(), createBlock(unref(ScalarSidebarSection), {
				key: 0,
				is: __props.is,
				class: "asyncapi-sidebar-filters"
			}, {
				items: withCtx(() => [showProtocol.value ? (openBlock(), createBlock(SidebarFilter_default, {
					key: 0,
					modelValue: protocol.value,
					"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => protocol.value = $event),
					label: "Protocol",
					options: protocolOptions.value
				}, null, 8, ["modelValue", "options"])) : createCommentVNode("", true), showServer.value ? (openBlock(), createBlock(SidebarFilter_default, {
					key: 1,
					modelValue: server.value,
					"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => server.value = $event),
					label: "Server",
					options: serverOptions.value
				}, null, 8, ["modelValue", "options"])) : createCommentVNode("", true)]),
				default: withCtx(() => [_cache[2] || (_cache[2] = createTextVNode(" Filters ", -1))]),
				_: 1
			}, 8, ["is"])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { AsyncApiSidebarFilters_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AsyncApiSidebarFilters.vue.script.js.map