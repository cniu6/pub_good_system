import { Fragment, computed, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconGavel } from "@scalar/icons";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
//#region src/features/info-object/License.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "group narrow:border-r-0 narrow:first:ml-0 flex h-fit items-center border-r first:ml-auto last:border-r-0" };
var _hoisted_2 = ["href"];
var _hoisted_3 = { class: "ml-1 empty:hidden" };
var _hoisted_4 = { class: "ml-1 empty:hidden" };
var License_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "License",
	props: { value: {} },
	setup(__props) {
		/**
		* The license URL comes from the OpenAPI document, which is untrusted input, so a protocol like
		* `javascript:` would execute script on click. Fall back to the plain text label in that case.
		*/
		const url = computed(() => sanitizeUrl(__props.value?.url));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [url.value ? (openBlock(), createElementBlock("a", {
				key: 0,
				class: "text-c-1 hover:bg-b-2 narrow:border mr-2 flex min-h-7 min-w-7 items-center rounded-lg px-2 py-1 no-underline group-last:mr-0",
				href: url.value,
				rel: "noopener noreferrer",
				target: "_blank"
			}, [createVNode(unref(ScalarIconGavel), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_3, toDisplayString(__props.value?.name || __props.value && "identifier" in __props.value && __props.value.identifier || url.value), 1)], 8, _hoisted_2)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createVNode(unref(ScalarIconGavel), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_4, toDisplayString(__props.value?.name), 1)], 64))]);
		};
	}
});
//#endregion
export { License_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=License.vue.script.js.map