import { usePluginManager } from "../../plugins/hooks/usePluginManager.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, mergeProps, openBlock, renderList, resolveDynamicComponent, unref, withCtx } from "vue";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
//#region src/features/specification-extension/SpecificationExtension.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "text-base"
};
var SpecificationExtension_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SpecificationExtension",
	props: { value: {} },
	setup(__props) {
		const { getSpecificationExtensions } = usePluginManager();
		/**
		* Extract registered OpenAPI extension names
		*/
		function getCustomExtensionNames(value) {
			return Object.keys(value ?? {}).filter((item) => item.startsWith("x-"));
		}
		/**
		* Get the components for the specification extensions
		*/
		function getCustomOpenApiExtensionComponents(extensionNames) {
			return extensionNames.flatMap((name) => getSpecificationExtensions(name)).filter((extension) => extension.component);
		}
		/**
		* Get the names of custom extensions from the provided value.
		*/
		const customExtensionNames = computed(() => getCustomExtensionNames(__props.value));
		/**
		* Get the components for the custom extensions.
		*/
		const customExtensions = computed(() => getCustomOpenApiExtensionComponents(customExtensionNames.value));
		return (_ctx, _cache) => {
			return typeof __props.value === "object" && customExtensions.value.length ? (openBlock(), createElementBlock("div", _hoisted_1, [(openBlock(true), createElementBlock(Fragment, null, renderList(customExtensions.value, (extension) => {
				return openBlock(), createBlock(unref(ScalarErrorBoundary), null, {
					default: withCtx(() => [extension.renderer ? (openBlock(), createBlock(resolveDynamicComponent(extension.renderer), mergeProps({
						key: 0,
						ref_for: true
					}, {
						[extension.name]: __props.value?.[extension.name],
						component: extension.component
					}), null, 16)) : (openBlock(), createBlock(resolveDynamicComponent(extension.component), mergeProps({
						key: 1,
						ref_for: true
					}, { [extension.name]: __props.value?.[extension.name] }), null, 16))]),
					_: 2
				}, 1024);
			}), 256))])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SpecificationExtension_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SpecificationExtension.vue.script.js.map