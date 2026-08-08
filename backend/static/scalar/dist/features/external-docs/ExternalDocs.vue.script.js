import { computed, createBlock, createCommentVNode, createElementBlock, createVNode, defineComponent, openBlock, resolveDynamicComponent, toDisplayString, unref, withCtx } from "vue";
import { ScalarIconBook } from "@scalar/icons";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
//#region src/features/external-docs/ExternalDocs.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0"
};
var _hoisted_2 = {
	key: 0,
	class: "ml-1 empty:hidden"
};
var _hoisted_3 = {
	key: 1,
	class: "ml-1 empty:hidden"
};
var ExternalDocs_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExternalDocs",
	props: { value: {} },
	setup(__props) {
		/**
		* The external docs URL comes from the OpenAPI document, which is untrusted input, so a protocol
		* like `javascript:` would execute script on click. Fall back to the plain text label in that case.
		*/
		const url = computed(() => sanitizeUrl(__props.value?.url));
		return (_ctx, _cache) => {
			return __props.value ? (openBlock(), createElementBlock("div", _hoisted_1, [(openBlock(), createBlock(resolveDynamicComponent(url.value ? "a" : "span"), {
				class: "text-c-1 hover:bg-b-2 narrow:border mr-2 flex min-h-7 min-w-7 items-center rounded-lg px-2 py-1 no-underline group-last:mr-0",
				href: url.value,
				rel: url.value ? "noopener noreferrer" : void 0,
				target: url.value ? "_blank" : void 0
			}, {
				default: withCtx(() => [createVNode(unref(ScalarIconBook), {
					class: "size-3 text-current",
					weight: "bold"
				}), __props.value.description ? (openBlock(), createElementBlock("span", _hoisted_2, toDisplayString(__props.value.description), 1)) : (openBlock(), createElementBlock("span", _hoisted_3, toDisplayString(__props.value.url), 1))]),
				_: 1
			}, 8, [
				"href",
				"rel",
				"target"
			]))])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { ExternalDocs_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExternalDocs.vue.script.js.map